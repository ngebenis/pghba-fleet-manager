-- ============================================================
-- SSH host key pinning
--
-- Every SSH connection this app makes previously used
-- ssh.InsecureIgnoreHostKey() — it never checked the remote host's SSH
-- key at all, so a network-level attacker able to intercept the
-- connection (ARP/DNS spoofing, a compromised router, etc.) could
-- man-in-the-middle it undetected. Store a trust-on-first-use
-- fingerprint per host: the first successful connection records it,
-- and every connection after that is refused if the key ever changes.
-- ============================================================

ALTER TABLE db_hosts ADD COLUMN IF NOT EXISTS ssh_host_key_fingerprint TEXT;
