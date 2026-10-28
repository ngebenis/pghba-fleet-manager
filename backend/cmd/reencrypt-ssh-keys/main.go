// reencrypt-ssh-keys re-encrypts every stored SSH private key (db_hosts.ssh_key)
// from an old AES-256 ENCRYPTION_KEY to a new one, without ever writing the
// decrypted plaintext anywhere but memory.
//
// Use this after rotating ENCRYPTION_KEY — e.g. because the key committed in
// backend/.env was published and must be treated as compromised. Run it
// BEFORE switching the running app over to the new key: this tool still
// needs the old key to decrypt what's currently stored.
//
// Usage:
//
//	OLD_ENCRYPTION_KEY=<64 hex chars> \
//	NEW_ENCRYPTION_KEY=<64 hex chars> \
//	DB_HOST=... DB_PORT=... DB_NAME=... DB_USER=... DB_PASSWORD=... DB_SSL_MODE=... \
//	go run ./cmd/reencrypt-ssh-keys [-dry-run]
//
// DB_* vars are the same ones the main app reads (backend/.env) — point this
// at the same database. -dry-run decrypts and re-encrypts in memory to
// verify every row round-trips cleanly, but writes nothing back, so you can
// sanity-check before committing to the real run.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/database"
	"github.com/pghba-manager/backend/internal/services"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "decrypt+re-encrypt every row to verify it round-trips, but write nothing back")
	flag.Parse()

	oldKeyHex := os.Getenv("OLD_ENCRYPTION_KEY")
	newKeyHex := os.Getenv("NEW_ENCRYPTION_KEY")
	if oldKeyHex == "" || newKeyHex == "" {
		log.Fatal("set both OLD_ENCRYPTION_KEY and NEW_ENCRYPTION_KEY (64 hex chars each) in the environment")
	}
	if oldKeyHex == newKeyHex {
		log.Fatal("OLD_ENCRYPTION_KEY and NEW_ENCRYPTION_KEY are identical — nothing to rotate")
	}

	oldCrypto, err := services.NewCryptoService(oldKeyHex)
	if err != nil {
		log.Fatalf("OLD_ENCRYPTION_KEY: %v", err)
	}
	newCrypto, err := services.NewCryptoService(newKeyHex)
	if err != nil {
		log.Fatalf("NEW_ENCRYPTION_KEY: %v", err)
	}

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	type row struct {
		ID     uuid.UUID
		Name   string
		SSHKey string
	}
	var rows []row
	if err := db.Table("db_hosts").Select("id, name, ssh_key").Where("ssh_key IS NOT NULL AND ssh_key != ''").Find(&rows).Error; err != nil {
		log.Fatalf("failed to load db_hosts: %v", err)
	}

	if len(rows) == 0 {
		fmt.Println("no hosts have an ssh_key set — nothing to do")
		return
	}

	mode := "LIVE"
	if *dryRun {
		mode = "DRY RUN (nothing will be written)"
	}
	fmt.Printf("Re-encrypting %d host(s) — %s\n", len(rows), mode)

	succeeded, failed := 0, 0
	for _, r := range rows {
		plaintext, err := oldCrypto.Decrypt(r.SSHKey)
		if err != nil {
			log.Printf("FAILED to decrypt host %q (%s) with OLD_ENCRYPTION_KEY: %v — leaving it untouched", r.Name, r.ID, err)
			failed++
			continue
		}

		reencrypted, err := newCrypto.Encrypt(plaintext)
		if err != nil {
			log.Printf("FAILED to re-encrypt host %q (%s) with NEW_ENCRYPTION_KEY: %v — leaving it untouched", r.Name, r.ID, err)
			failed++
			continue
		}

		// Round-trip check before writing anything: decrypt what we're
		// about to store and confirm it matches the original plaintext
		// byte for byte, so a subtle bug here can never corrupt a key.
		roundTrip, err := newCrypto.Decrypt(reencrypted)
		if err != nil || roundTrip != plaintext {
			log.Printf("FAILED round-trip check for host %q (%s) — leaving it untouched", r.Name, r.ID)
			failed++
			continue
		}

		if *dryRun {
			fmt.Printf("  OK (dry run)  %-30s %s\n", r.Name, r.ID)
			succeeded++
			continue
		}

		if err := db.Table("db_hosts").Where("id = ?", r.ID).Update("ssh_key", reencrypted).Error; err != nil {
			log.Printf("FAILED to write re-encrypted key for host %q (%s): %v", r.Name, r.ID, err)
			failed++
			continue
		}
		fmt.Printf("  OK            %-30s %s\n", r.Name, r.ID)
		succeeded++
	}

	fmt.Printf("\nDone: %d succeeded, %d failed.\n", succeeded, failed)
	if failed > 0 {
		fmt.Println("Any failed host still has its OLD key stored — fix the issue and re-run before switching the app to NEW_ENCRYPTION_KEY, or those hosts will become undecryptable.")
		os.Exit(1)
	}
	if !*dryRun {
		fmt.Println("All SSH keys re-encrypted. You can now set ENCRYPTION_KEY=$NEW_ENCRYPTION_KEY in the app's environment and restart it.")
	}
}
