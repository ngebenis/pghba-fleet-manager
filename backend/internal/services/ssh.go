package services

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type SSHClient struct {
	client *ssh.Client
	user   string
}

// dialSSHVerified pins the remote host key by SHA256 fingerprint instead
// of trusting whatever key the network hands back on every call
// (ssh.InsecureIgnoreHostKey, the previous behavior, has zero protection
// against a network-level man-in-the-middle). expectedFingerprint == ""
// means trust-on-first-use: the connection is allowed and its fingerprint
// is returned for the caller to persist; once a fingerprint is on file,
// a mismatch is refused rather than silently accepted.
func dialSSHVerified(host string, port int, user string, privateKeyPEM string, expectedFingerprint string) (*SSHClient, string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse private key: %w", err)
	}
	var seenFingerprint string
	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			seenFingerprint = ssh.FingerprintSHA256(key)
			if expectedFingerprint != "" && seenFingerprint != expectedFingerprint {
				return fmt.Errorf(
					"host key for %s does not match the pinned fingerprint (expected %s, got %s) — "+
						"this could mean the host legitimately rebuilt/rotated its key, or a man-in-the-middle; "+
						"verify out of band before clearing the stored fingerprint for this host",
					hostname, expectedFingerprint, seenFingerprint)
			}
			return nil
		},
		Timeout: 10 * time.Second,
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, seenFingerprint, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	return &SSHClient{client: client, user: user}, seenFingerprint, nil
}

// ConnectToHost is the normal way to open an SSH connection to a managed
// host: it pins the host's SSH key by fingerprint (dialSSHVerified) and,
// on an empty storedFingerprint (first-ever connection to this host),
// persists the fingerprint it learns so every later connection can be
// checked against it.
func ConnectToHost(db *gorm.DB, hostID uuid.UUID, host string, port int, user string, privateKeyPEM string, storedFingerprint string) (*SSHClient, error) {
	client, fp, err := dialSSHVerified(host, port, user, privateKeyPEM, storedFingerprint)
	if err != nil {
		return nil, err
	}
	if storedFingerprint == "" && fp != "" {
		db.Table("db_hosts").Where("id = ?", hostID).Update("ssh_host_key_fingerprint", fp)
	}
	return client, nil
}

// User returns the remote SSH username of this session.
func (s *SSHClient) User() string { return s.user }

// IsPostgresUser reports whether this session already runs as the
// postgres OS user.
func (s *SSHClient) IsPostgresUser() bool { return s.user == "postgres" }

func (s *SSHClient) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// shellQuote wraps s in single quotes for safe embedding in a remote shell
// command. Blocks the only character that can escape single quotes (')
// itself; callers must reject newlines/NUL separately if relevant.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// ShellQuotePath quotes a remote file path for shell embedding (exported
// for handlers; same rule as shellQuote).
func ShellQuotePath(s string) string { return shellQuote(s) }

// AsPostgres wraps cmd so it runs as the postgres OS user when the SSH
// session user is NOT postgres (via sudo -n, non-interactive — never
// prompts for a password). When already connected as postgres it returns
// cmd unchanged.
//
// Rationale: systemd-logind with the default RemoveIPC=yes destroys the
// postgres user's System V IPC / shm segments every time the last postgres
// SSH session logs out. Probing and managing 14 hosts as user postgres
// therefore risks wiping shared memory on every check cycle (incident
// 2026-10-23: all 4 clusters on 10.1.3.91 lost /dev/shm segments after
// loginctl "Stopping User Manager for UID 1002"). SSH should connect as a
// non-postgres ops user (e.g. dco) granted NOPASSWD sudo for postgres, and
// only PostgreSQL-touching commands escalate via this helper. Pure-OS
// probes (uptime, echo) keep using RunCommand directly.
func (s *SSHClient) AsPostgres(cmd string) string {
	if s.IsPostgresUser() {
		return cmd
	}
	return "sudo -n -u postgres bash -c " + shellQuote(cmd)
}

// RunAsPostgres runs cmd as the postgres OS user (see AsPostgres).
func (s *SSHClient) RunAsPostgres(cmd string) (string, error) {
	return s.RunCommand(s.AsPostgres(cmd))
}

// RunAsPostgresWithInput is RunCommandWithInput as the postgres OS user
// (see AsPostgres). Used for atomic file writes of postgres-owned files.
func (s *SSHClient) RunAsPostgresWithInput(cmd string, input string) (string, error) {
	return s.RunCommandWithInput(s.AsPostgres(cmd), input)
}

// RequireNonPostgresUser returns an error when the session runs as the
// postgres OS user. Call at the top of handlers/services that open an SSH
// session per request or per scheduler tick, to fail fast instead of
// risking IPC cleanup on session logout.
func (s *SSHClient) RequireNonPostgresUser() error {
	if s.IsPostgresUser() {
		return fmt.Errorf("refusing to run as OS user postgres: use a non-postgres SSH user (e.g. dco) with sudo -n -u postgres; postgres logins trigger systemd-logind IPC cleanup (RemoveIPC=yes) on logout")
	}
	return nil
}

func (s *SSHClient) RunCommand(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("command failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunCommandWithInput runs cmd on the remote host, feeding `input` on stdin.
// Use this instead of interpolating arbitrary content into a shell command
// string (heredoc/echo) — stdin has no quoting or terminator-collision
// issues regardless of what the content contains.
func (s *SSHClient) RunCommandWithInput(cmd string, input string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	session.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	session.Stdout = &out
	session.Stderr = &out
	if err := session.Run(cmd); err != nil {
		return out.String(), fmt.Errorf("command failed: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// TestSSHConnection dials host:port as user with the given private key and
// runs a harmless command to confirm the session actually works (a bare TCP
// connect can succeed against a non-SSH listener). Used as a pre-flight
// check from the Add/Edit Host form, before any host row is saved — so
// errors must stay descriptive enough to show directly in the UI rather
// than the raw crypto/ssh error text.
//
// Deliberately not pinned like ConnectToHost: this runs against whatever
// IP/port the form currently holds, which may not match a saved host (or
// any host) yet, and it never runs a real command against the host — it's
// a one-shot "can I reach this" probe the user triggers explicitly, not an
// operational channel worth persisting a fingerprint for.
func TestSSHConnection(host string, port int, user string, privateKeyPEM string, timeout time.Duration) error {
	signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
	if err != nil {
		return fmt.Errorf("invalid SSH private key: %v", err)
	}
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return classifySSHDialError(addr, user, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("connected to %s but could not open a session: %v", addr, err)
	}
	defer session.Close()
	if out, err := session.CombinedOutput("true"); err != nil {
		return fmt.Errorf("connected to %s as %s but command execution failed: %v (%s)", addr, user, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// classifySSHDialError turns the underlying net/ssh error into a message
// that names the likely cause instead of a bare "dial tcp ...: ...".
func classifySSHDialError(addr, user string, err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no route to host"):
		return fmt.Errorf("no route to host %s — check network routing/firewall between the app and this host", addr)
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("connection refused on %s — SSH may not be running there, or the port is wrong", addr)
	case strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "timed out"):
		return fmt.Errorf("connection to %s timed out — host may be unreachable or a firewall is silently dropping packets", addr)
	case strings.Contains(msg, "unable to authenticate") || strings.Contains(msg, "no supported methods remain"):
		return fmt.Errorf("permission denied for user %q on %s — the SSH key is not authorized for this user (check authorized_keys / Deploy SSH Key)", user, addr)
	case strings.Contains(msg, "no such host") || strings.Contains(msg, "server misbehaving") || strings.Contains(msg, "lookup"):
		return fmt.Errorf("could not resolve host %s — check the IP address/hostname", addr)
	default:
		return fmt.Errorf("SSH connection to %s failed: %v", addr, err)
	}
}

func CheckTCPPort(host string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// PGClusterInfo hasil parse dari pg_lsclusters atau pg16-lsclusters
type PGClusterInfo struct {
	Version     string
	ClusterName string
	Port        int
	Status      string
	DataDir     string
	HBAPath     string
}

// DiscoverAllClusters jalankan semua lsclusters commands yang tersedia
// dan gabungkan hasilnya. Cluster discovery membaca file config milik
// postgres (/etc/postgresql, postgresql.conf) — selalu via RunAsPostgres
// agar bisa jalan dari SSH user non-postgres. Versi/cluster yang
// didapat dari nama direktori di-quote agar aman walau namanya aneh.
func (s *SSHClient) DiscoverAllClusters() ([]PGClusterInfo, error) {
	var all []PGClusterInfo

	// 1. Standard pg_lsclusters (PG 10, 12, 14)
	out, err := s.RunAsPostgres("pg_lsclusters --no-header 2>/dev/null")
	if err == nil && strings.TrimSpace(out) != "" {
		clusters := ParseStandardLsClusters(out)
		all = append(all, clusters...)
	}

	// 2. Cari semua pg{ver}-lsclusters yang tersedia (PG 16, 17, 18, dst)
	// List binary yang ada di PATH
	binOut, _ := s.RunAsPostgres("ls /usr/bin/pg*-lsclusters /usr/local/bin/pg*-lsclusters 2>/dev/null")
	for _, bin := range strings.Split(binOut, "\n") {
		bin = strings.TrimSpace(bin)
		if bin == "" {
			continue
		}
		if strings.ContainsAny(bin, " 	\n'\";|&$`()\\") {
			continue // nama binary tak wajar — skip daripada dieksekusi
		}
		// Jalankan binary tersebut
		customOut, err := s.RunAsPostgres(bin + " 2>/dev/null")
		if err != nil || strings.TrimSpace(customOut) == "" {
			continue
		}
		// Parse format custom (pg16-lsclusters)
		clusters := ParseCustomLsClusters(customOut)
		all = append(all, clusters...)
	}

	// 3. Fallback: cari manual di /etc/postgresql dan /opt/postgresql
	if len(all) == 0 {
		pgDirs, _ := s.RunAsPostgres("ls /etc/postgresql/ 2>/dev/null")
		for _, ver := range strings.Split(pgDirs, "\n") {
			ver = strings.TrimSpace(ver)
			if ver == "" {
				continue
			}
			clusters, _ := s.RunAsPostgres(fmt.Sprintf(
				"ls /etc/postgresql/%s/ 2>/dev/null", shellQuote(ver),
			))
			for _, cl := range strings.Split(clusters, "\n") {
				cl = strings.TrimSpace(cl)
				if cl == "" {
					continue
				}
				// Baca port dari postgresql.conf
				portOut, _ := s.RunAsPostgres(fmt.Sprintf(
					"grep -E '^port' /etc/postgresql/%s/%s/postgresql.conf 2>/dev/null | awk '{print $3}'",
					shellQuote(ver), shellQuote(cl),
				))
				var port int
				fmt.Sscanf(strings.TrimSpace(portOut), "%d", &port)
				if port == 0 {
					continue
				}
				dataOut, _ := s.RunAsPostgres(fmt.Sprintf(
					"grep -E '^data_directory' /etc/postgresql/%s/%s/postgresql.conf 2>/dev/null | awk -F\"'\" '{print $2}'",
					shellQuote(ver), shellQuote(cl),
				))
				dataDir := strings.TrimSpace(dataOut)
				if dataDir == "" {
					dataDir = fmt.Sprintf("/var/lib/postgresql/%s/%s", ver, cl)
				}
				all = append(all, PGClusterInfo{
					Version:     ver,
					ClusterName: cl,
					Port:        port,
					Status:      "unknown",
					DataDir:     dataDir,
					HBAPath:     fmt.Sprintf("/etc/postgresql/%s/%s/pg_hba.conf", ver, cl),
				})
			}
		}
	}

	return all, nil
}

// ParseStandardLsClusters parse output pg_lsclusters --no-header
// Format: Ver Cluster Port Status Owner DataDirectory LogFile
func ParseStandardLsClusters(output string) []PGClusterInfo {
	var clusters []PGClusterInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Ver") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		var port int
		fmt.Sscanf(fields[2], "%d", &port)
		if port == 0 {
			continue
		}
		ver := fields[0]
		clusterName := fields[1]
		status := fields[3]
		dataDir := fields[5]

		clusters = append(clusters, PGClusterInfo{
			Version:     ver,
			ClusterName: clusterName,
			Port:        port,
			Status:      status,
			DataDir:     dataDir,
			HBAPath:     HBAPathFromCluster(ver, clusterName),
		})
	}
	return clusters
}

// ParseCustomLsClusters parse output pg16-lsclusters
// Format:
// PostgreSQL 16 Manual Clusters:
// ===============================================
// Instance       Port   Status    Data Directory
// -----------------------------------------------
// appsdb           5442   online    /opt/postgresql/16new/data
// cperpduckpir       5438   online    /opt/postgresql/16new/cperpduckpir/data
func ParseCustomLsClusters(output string) []PGClusterInfo {
	var clusters []PGClusterInfo
	var pgVersion string

	// Deteksi versi dari header "PostgreSQL 16 Manual Clusters:"
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PostgreSQL") && strings.Contains(line, "Manual Clusters") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pgVersion = parts[1]
			}
			break
		}
	}

	if pgVersion == "" {
		return clusters
	}

	// Parse baris data (skip header, separator, dan baris Services)
	inData := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		// Mulai baca data setelah baris separator kedua (----)
		if strings.HasPrefix(line, "---") {
			inData = true
			continue
		}
		// Stop saat ketemu "Services:" atau baris kosong setelah data
		if strings.HasPrefix(line, "Services:") || strings.HasPrefix(line, "Active:") {
			inData = false
			continue
		}
		// Skip header row
		if strings.HasPrefix(line, "Instance") {
			continue
		}

		if !inData || line == "" {
			continue
		}

		// Parse: InstanceName Port Status DataDirectory
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		clusterName := fields[0]
		var port int
		fmt.Sscanf(fields[1], "%d", &port)
		if port == 0 {
			continue
		}
		status := fields[2]
		dataDir := fields[3]

		// pg_hba.conf biasanya ada di data directory
		hbaPath := dataDir + "/pg_hba.conf"

		clusters = append(clusters, PGClusterInfo{
			Version:     pgVersion,
			ClusterName: clusterName,
			Port:        port,
			Status:      status,
			DataDir:     dataDir,
			HBAPath:     hbaPath,
		})
	}
	return clusters
}

// HBAPathFromCluster derive pg_hba.conf path standard Debian/Ubuntu
func HBAPathFromCluster(version, clusterName string) string {
	return fmt.Sprintf("/etc/postgresql/%s/%s/pg_hba.conf", version, clusterName)
}
