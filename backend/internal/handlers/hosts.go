package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type HostHandler struct {
	db     *gorm.DB
	audit  *services.AuditService
	crypto *services.CryptoService
}

func NewHostHandler(db *gorm.DB, audit *services.AuditService, crypto *services.CryptoService) *HostHandler {
	return &HostHandler{db: db, audit: audit, crypto: crypto}
}

func (h *HostHandler) List(c *gin.Context) {
	var hosts []models.DBHost
	h.db.Find(&hosts)

	type HostWithInstances struct {
		models.DBHost
		InstanceCount int64               `json:"instance_count"`
		Instances     []models.PGInstance `json:"instances"`
	}

	result := make([]HostWithInstances, len(hosts))
	for i, host := range hosts {
		result[i].DBHost = host
		h.db.Where("host_id = ?", host.ID).Find(&result[i].Instances)
		result[i].InstanceCount = int64(len(result[i].Instances))
		if result[i].SSHKey != "" {
			result[i].SSHKey = "••••••••"
		}
	}
	c.JSON(http.StatusOK, result)
}

func (h *HostHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	host.SSHKey = ""
	c.JSON(http.StatusOK, host)
}

func (h *HostHandler) Create(c *gin.Context) {
	var host models.DBHost
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encrypt SSH key if provided
	if host.SSHKey != "" && host.SSHKey != "••••••••" {
		enc, err := h.crypto.Encrypt(host.SSHKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt SSH key"})
			return
		}
		host.SSHKey = enc
	} else if host.SSHKey == "" {
		// Auto-inherit SSH key from existing host
		var existing models.DBHost
		if err := h.db.Where("ssh_key != ''").First(&existing).Error; err == nil {
			host.SSHKey = existing.SSHKey
		}
	}

	if err := h.db.Create(&host).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventAddHost,
		host.Name, nil, "", c.ClientIP(),
		map[string]interface{}{"ip": host.IPAddress, "env": host.Environment})

	host.SSHKey = ""
	c.JSON(http.StatusCreated, host)
}

func (h *HostHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)

	// The pinned SSH host key fingerprint is security state, not an
	// editable field — never let a generic update silently change or
	// clear it (that would let a compromised session re-pin to an
	// attacker's key after a MITM, defeating ConnectToHost's whole point).
	// Resetting it after a legitimate host key rotation goes through the
	// dedicated ResetHostKeyFingerprint endpoint instead, so it's a
	// deliberate, audited action.
	delete(updates, "ssh_host_key_fingerprint")

	if key, ok := updates["ssh_key"].(string); ok && key != "" && key != "••••••••" {
		enc, _ := h.crypto.Encrypt(key)
		updates["ssh_key"] = enc
	} else {
		delete(updates, "ssh_key")
	}

	h.db.Model(&host).Updates(updates)
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditHost,
		host.Name, nil, "", c.ClientIP(), nil)
	host.SSHKey = ""
	c.JSON(http.StatusOK, host)
}

// ResetHostKeyFingerprint clears the pinned SSH host key fingerprint for a
// host, so the next connection trusts whatever key it sees and re-pins
// from there. Needed after a legitimate host key rotation (host rebuilt,
// OS reinstalled) — without this, ConnectToHost would permanently refuse
// to connect to that host, mistaking the new key for a MITM. Deliberately
// a separate, audited action rather than part of the generic Update, so
// clearing it is never silent.
func (h *HostHandler) ResetHostKeyFingerprint(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	h.db.Model(&host).Update("ssh_host_key_fingerprint", "")
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditHost,
		host.Name+" (reset SSH host key fingerprint)", nil, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "host key fingerprint cleared — the next connection will pin a new one"})
}

// Validate runs an SSH pre-flight check against the given connection
// details without saving anything. The Add/Edit Host form calls this
// before submitting so a bad IP, closed port, firewall, or unauthorized
// key is caught with a descriptive message instead of discovered later
// when a write/reload silently fails.
func (h *HostHandler) Validate(c *gin.Context) {
	var r struct {
		HostID    *uuid.UUID `json:"host_id"` // set when editing an existing host
		IPAddress string     `json:"ip_address" binding:"required"`
		SSHPort   int        `json:"ssh_port"`
		SSHUser   string     `json:"ssh_user"`
		SSHKey    string     `json:"ssh_key"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if r.SSHPort == 0 {
		r.SSHPort = 22
	}
	if r.SSHUser == "" {
		r.SSHUser = "dco"
	}

	privateKey, err := h.resolveSSHKeyForValidate(r.HostID, r.SSHKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if err := services.TestSSHConnection(r.IPAddress, r.SSHPort, r.SSHUser, privateKey, 5*time.Second); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": fmt.Sprintf("SSH connection to %s:%d as %s succeeded", r.IPAddress, r.SSHPort, r.SSHUser),
	})
}

// resolveSSHKeyForValidate mirrors the key-inheritance rule used by Create:
// an explicit key wins; otherwise fall back to the host being edited, then
// to any existing host's key, so validation reflects the key that will
// actually end up being used.
func (h *HostHandler) resolveSSHKeyForValidate(hostID *uuid.UUID, explicitKey string) (string, error) {
	if explicitKey != "" && explicitKey != "••••••••" {
		return explicitKey, nil
	}

	var stored models.DBHost
	if hostID != nil {
		if err := h.db.First(&stored, "id = ?", *hostID).Error; err == nil && stored.SSHKey != "" {
			return h.crypto.Decrypt(stored.SSHKey)
		}
	}
	if err := h.db.Where("ssh_key != ''").First(&stored).Error; err == nil {
		return h.crypto.Decrypt(stored.SSHKey)
	}
	return "", fmt.Errorf("no SSH key provided and no existing SSH key found to inherit — paste a private key first")
}

func (h *HostHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	claims := middleware.GetClaims(c)
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.db.Delete(&host)
	h.audit.Log(&claims.UserID, claims.Username, models.EventDeleteHost,
		host.Name, nil, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "host deleted"})
}

// DeploySSHKey — copy public key to target host via password SSH, then
// grant the ops user passwordless sudo to postgres.
//
// New model (post 2026-10-23 IPC incident): the app must NEVER log in as
// the postgres OS user — every postgres login risks systemd-logind
// (RemoveIPC=yes) wiping postgres IPC/shm on logout. So this endpoint now
// deploys the key to an OPS user (default: host.SSHUser when it is not
// postgres, else "dco") and writes a restricted sudoers drop-in:
//
//	<opsuser> ALL=(postgres) NOPASSWD: /bin/bash -c *
//
// Deployment to user postgres is refused unless the caller explicitly
// passes "pg_user": "postgres" (legacy mode, discouraged).
func (h *HostHandler) DeploySSHKey(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))

	type req struct {
		RootPassword string `json:"root_password" binding:"required"`
		SSHUser      string `json:"ssh_user"`
		OpsUser      string `json:"ops_user"`
		PGUser       string `json:"pg_user"` // legacy override; default refused for postgres
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load host
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	if r.SSHUser == "" {
		r.SSHUser = "root"
	}
	// Ops user = explicit ops_user, else host.SSHUser when it is not
	// postgres, else "dco". The app never logs in as postgres anymore.
	opsUser := r.OpsUser
	if opsUser == "" {
		opsUser = r.PGUser // backward compat: old callers sent target in pg_user
	}
	if opsUser == "" {
		opsUser = host.SSHUser
	}
	if opsUser == "" || opsUser == "postgres" || opsUser == "root" {
		if r.PGUser == "postgres" && r.OpsUser == "" {
			// Legacy explicit mode preserved below (discouraged).
		} else {
			opsUser = "dco"
		}
	}

	// Get public key from any existing host
	var existingHost models.DBHost
	if err := h.db.Where("ssh_key != ''").First(&existingHost).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no SSH key found in system"})
		return
	}

	privateKey, err := h.crypto.Decrypt(existingHost.SSHKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt SSH key"})
		return
	}

	// Parse private key to get public key
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse private key"})
		return
	}
	pubKey := string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
	pubKey = strings.TrimSpace(pubKey)

	// Connect to target host as root using password
	sshPort := host.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}

	rootClient, fp, err := connectWithPassword(host.IPAddress, sshPort, r.SSHUser, r.RootPassword, host.SSHHostKeyFingerprint)
	if err == nil && host.SSHHostKeyFingerprint == "" && fp != "" {
		h.db.Model(&host).Update("ssh_host_key_fingerprint", fp)
		host.SSHHostKeyFingerprint = fp
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("SSH connection failed as %s: %s", r.SSHUser, err.Error()),
		})
		return
	}
	defer rootClient.Close()

	// Run commands to authorize the key for the OPS user (legacy mode:
	// r.PGUser=="postgres" keeps old postgres target). Note: with sudo -n
	// the ops user keeps no postgres login session, so no IPC cleanup.
	targetUser := opsUser
	legacyPostgres := r.PGUser == "postgres" && r.OpsUser == ""
	if legacyPostgres {
		targetUser = "postgres"
	}
	targetHomeCmd := fmt.Sprintf("getent passwd %s | cut -d: -f6", targetUser)
	targetHome, err := runSSHCommand(rootClient, targetHomeCmd)
	if err != nil || strings.TrimSpace(targetHome) == "" {
		if legacyPostgres {
			targetHome = "/var/lib/postgresql"
		} else {
			targetHome = "/home/" + targetUser
		}
	}
	targetHome = strings.TrimSpace(targetHome)

	sshDir := targetHome + "/.ssh"
	authFile := sshDir + "/authorized_keys"

	cmds := []string{
		fmt.Sprintf("mkdir -p %s", sshDir),
		fmt.Sprintf("touch %s", authFile),
		fmt.Sprintf("grep -qF '%s' %s || echo '%s' >> %s", pubKey, authFile, pubKey, authFile),
		fmt.Sprintf("chown -R %s:%s %s", targetUser, targetUser, sshDir),
		fmt.Sprintf("chmod 700 %s", sshDir),
		fmt.Sprintf("chmod 600 %s", authFile),
	}

	// Restricted sudoers drop-in: ops user may run postgres commands
	// non-interactively (skipped in legacy postgres mode).
	sudoersEntry := ""
	if !legacyPostgres {
		sudoersEntry = fmt.Sprintf("%s ALL=(postgres) NOPASSWD: /bin/bash -c *", targetUser)
		sudoersFile := fmt.Sprintf("/etc/sudoers.d/pghba-%s", targetUser)
		cmds = append(cmds,
			fmt.Sprintf("printf '%%s\\n' '%s' > %s && chmod 440 %s && visudo -c -f %s",
				sudoersEntry, sudoersFile, sudoersFile, sudoersFile),
		)
	}

	results := []string{}
	for _, cmd := range cmds {
		out, err := runSSHCommand(rootClient, cmd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   fmt.Sprintf("command failed: %s", cmd),
				"detail":  err.Error(),
				"output":  out,
				"results": results,
			})
			return
		}
		results = append(results, fmt.Sprintf("✓ %s", cmd))
	}

	// Test SSH as ops user; pg_lsclusters via sudo -n (proves the exact
	// privilege path the app will use).
	opsClient, testErr := services.ConnectToHost(h.db, host.ID, host.IPAddress, sshPort, targetUser, privateKey, host.SSHHostKeyFingerprint)
	testResult := ""
	if testErr != nil {
		testResult = fmt.Sprintf("⚠ SSH test as %s failed: %s", targetUser, testErr.Error())
	} else {
		out, _ := opsClient.RunCommand("sudo -n -u postgres bash -c 'pg_lsclusters --no-header 2>/dev/null' || echo 'connected-no-pg_lsclusters'")
		opsClient.Close()
		testResult = fmt.Sprintf("✓ SSH test as %s succeeded: %s", targetUser, strings.TrimSpace(out))
	}

	// Update host SSH key if not already set; point ssh_user at the ops
	// user so future connections never log in as postgres.
	if host.SSHKey == "" {
		h.db.Model(&host).Update("ssh_key", existingHost.SSHKey)
	}
	if !legacyPostgres && host.SSHUser != targetUser {
		h.db.Model(&host).Update("ssh_user", targetUser)
	}

	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditHost,
		host.Name+" (deploy SSH key)", nil, "", c.ClientIP(),
		map[string]interface{}{"ip": host.IPAddress, "ops_user": targetUser, "sudoers": sudoersEntry})

	c.JSON(http.StatusOK, gin.H{
		"message":     "SSH key deployed successfully",
		"host":        host.Name,
		"ip":          host.IPAddress,
		"ops_user":    targetUser,
		"sudoers":     sudoersEntry,
		"auth_file":   authFile,
		"commands":    results,
		"test_result": testResult,
		"ready":       testErr == nil,
	})
}

// connectWithPassword connects via SSH using password auth. This is
// normally the very first SSH contact ever made with a brand new host
// (before any key — and so any pinned fingerprint — exists for it), so it
// pins the host key the same way services.ConnectToHost does: trust on
// first use when expectedFingerprint is empty, refuse on any later
// mismatch. Returns the fingerprint actually seen so the caller can
// persist it.
func connectWithPassword(host string, port int, user, password string, expectedFingerprint string) (*ssh.Client, string, error) {
	var seenFingerprint string
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = password
				}
				return answers, nil
			}),
		},
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
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, config)
	return client, seenFingerprint, err
}

// runSSHCommand runs a command on existing SSH connection
func runSSHCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)
	return string(out), err
}
