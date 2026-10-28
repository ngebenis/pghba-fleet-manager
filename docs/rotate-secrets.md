# Rotating JWT_SECRET and ENCRYPTION_KEY

`backend/.env` — containing the real `JWT_SECRET` and `ENCRYPTION_KEY` — was
committed to this repository from the very first commit. Both must be
treated as compromised before this repo is made public:

- **`JWT_SECRET`** signs every login token. Anyone who reads the source
  history can forge a valid token for any user on a deployment that still
  uses it.
- **`ENCRYPTION_KEY`** is the AES-256 key for every SSH private key stored
  in `db_hosts.ssh_key`. Anyone who reads the source history and later gets
  read access to the database (a backup, a dump, a different vulnerability)
  can decrypt every managed host's SSH key.

This runbook rotates both. Run it against your production deployment —
nothing here can be run from this session, since it has no access to that
database or server.

## 1. Generate new secrets

```bash
openssl rand -base64 48   # new JWT_SECRET
openssl rand -hex 32      # new ENCRYPTION_KEY (must be exactly 64 hex chars)
```

Save both somewhere safe for the next steps. Do not reuse the values
currently in `backend/.env` — those are the compromised ones.

## 2. Re-encrypt every stored SSH key

`JWT_SECRET` needs no migration (it only signs tokens, nothing is encrypted
with it) — just set the new value in step 4. `ENCRYPTION_KEY` is different:
every row in `db_hosts.ssh_key` is AES-256-encrypted with the *current* key,
so changing it without re-encrypting first makes every stored SSH key
permanently undecryptable.

A dedicated tool handles this: `backend/cmd/reencrypt-ssh-keys`. Build and
run it on a host that can reach your production database directly — from
the `backend/` directory of this repo, on the same machine/network as the
production app or via an SSH tunnel to the database:

```bash
cd backend

# Dry run first — decrypts every key with the OLD key and re-encrypts it
# with the NEW key in memory only, verifies the round-trip, writes nothing.
OLD_ENCRYPTION_KEY=<the current value from backend/.env>   \
NEW_ENCRYPTION_KEY=<the new value from step 1>              \
DB_HOST=<production DB host> DB_PORT=5432 DB_NAME=pghba_manager \
DB_USER=<production DB user> DB_PASSWORD=<production DB password> DB_SSL_MODE=<disable|require> \
go run ./cmd/reencrypt-ssh-keys -dry-run

# Once the dry run reports "0 failed" for every host, run it for real
# (same command, drop -dry-run):
OLD_ENCRYPTION_KEY=<...> NEW_ENCRYPTION_KEY=<...> DB_HOST=<...> DB_PORT=5432 \
DB_NAME=pghba_manager DB_USER=<...> DB_PASSWORD=<...> DB_SSL_MODE=<...> \
go run ./cmd/reencrypt-ssh-keys
```

Do this while the app isn't actively writing to `db_hosts.ssh_key` (no one
editing a host's SSH key or running "Deploy SSH Key" at the same time) —
the tool reads, decrypts, and writes back per row, so a concurrent write to
the same row could race with it.

If the tool reports any failures, it leaves those rows on the OLD key and
tells you so — fix whatever's wrong (usually a copy-paste error in
`OLD_ENCRYPTION_KEY`) and re-run before continuing; don't switch the app to
the new key while any host is still on the old one.

## 3. Rotate the database password too (recommended, not strictly required)

`DB_PASSWORD` in `backend/.env` is also committed, though it's lower risk
since that database normally isn't reachable outside your own network.
Still worth changing it in the same maintenance window:

```sql
ALTER USER <db_user> WITH PASSWORD '<new password>';
```

## 4. Update the running app's environment and restart

Edit the production `backend/.env` (or however the deployment supplies
environment variables — e.g. `docker-compose.yml`'s `environment:` block,
a secrets manager) and set:

```
JWT_SECRET=<new value from step 1>
ENCRYPTION_KEY=<new value from step 1>
DB_PASSWORD=<new value from step 3, if rotated>
```

Then restart the app (`docker compose restart app`, or equivalent). The
server now refuses to start if either `JWT_SECRET` or `ENCRYPTION_KEY` is
left at the published default (see `config.Validate()` in `main.go`) — if
you made a typo, you'll find out immediately rather than running insecurely.

**Everyone currently logged in is signed out** — the old `JWT_SECRET` no
longer validates any existing token, so every user has to log in again.
This is expected and fine; mention it to your team beforehand if it's a
shared deployment.

## 5. Confirm

- Log in — should work normally with a fresh token.
- Open an existing host and run "Test" (or trigger a scan/reload on one of
  its instances) — if it connects successfully, its SSH key decrypted
  correctly under the new `ENCRYPTION_KEY`.
- Check the app's logs for any `decrypt SSH key failed` errors on startup
  or first use — that would mean a host's key didn't get re-encrypted in
  step 2.

## 6. Publishing this repo

This runbook only fixes the *running deployment*. The compromised secrets
are also sitting in `backend/.env` inside this repository's git history
since the first commit — rotating the live secrets does not remove them
from history. Per your decision, publish this codebase as a **new**
repository built from a fresh snapshot (no old commit history) rather than
rewriting this repo's history in place.
