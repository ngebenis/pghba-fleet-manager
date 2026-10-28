package services

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/models"
)

// ParseHBAContent parse isi pg_hba.conf dan return slice HBARule
func ParseHBAContent(content string, instanceID uuid.UUID) []models.HBARule {
	var rules []models.HBARule

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	for i, line := range strings.Split(content, "\n") {
		original := line
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		enabled := true

		// Cek disabled rule (diawali #)
		if strings.HasPrefix(line, "#") {
			stripped := strings.TrimSpace(line[1:])
			if !startsWithRuleType(stripped) {
				continue
			}
			line = stripped
			enabled = false
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		ruleType := models.RuleType(strings.ToLower(fields[0]))

		validTypes := map[models.RuleType]bool{
			"local":       true,
			"host":        true,
			"hostssl":     true,
			"hostnossl":   true,
			"hostgssenc":  true,
			"hostnogssenc": true,
		}
		if !validTypes[ruleType] {
			continue
		}

		rule := models.HBARule{
			InstanceID: instanceID,
			RuleType:   ruleType,
			Enabled:    enabled,
			RuleOrder:  i,
		}

		if ruleType == "local" {
			// local DATABASE USER METHOD [OPTIONS]
			if len(fields) < 4 {
				continue
			}
			rule.Database   = fields[1]
			rule.DBUser     = fields[2]
			rule.AuthMethod = models.AuthMethod(fields[3])
		} else {
			// host/hostssl/etc DATABASE USER ADDRESS METHOD [OPTIONS]
			if len(fields) < 5 {
				continue
			}
			rule.Database   = fields[1]
			rule.DBUser     = fields[2]
			rule.Address    = fields[3]
			rule.AuthMethod = models.AuthMethod(fields[4])
		}

		// Validasi AuthMethod tidak kosong
		if rule.AuthMethod == "" {
			continue
		}

		// Ambil inline comment
		if idx := strings.Index(original, "#"); idx > 0 {
			candidate := strings.TrimSpace(original[idx+1:])
			if candidate != "" && !startsWithRuleType(candidate) {
				rule.Comment = candidate
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

// FindUntrackedLiveRules parses the pg_hba.conf content currently on the
// server and returns a human-readable line for each ENABLED rule that has
// no matching rule (by type/database/user/address — regardless of that
// rule's enabled state) anywhere in dbRules.
//
// A non-empty result means the live file has active rules the database
// doesn't know about — e.g. added by hand on the server, or never
// imported. Since every write-back regenerates the entire file from the
// database, writing now would silently delete those rules. Callers should
// refuse to write and surface this instead.
func FindUntrackedLiveRules(currentContent string, instanceID uuid.UUID, dbRules []models.HBARule) []string {
	if strings.TrimSpace(currentContent) == "" {
		return nil
	}

	known := make(map[string]bool, len(dbRules))
	for _, r := range dbRules {
		known[ruleSignature(r)] = true
	}

	var missing []string
	for _, lr := range ParseHBAContent(currentContent, instanceID) {
		if !lr.Enabled {
			continue
		}
		if !known[ruleSignature(lr)] {
			missing = append(missing, lr.ToHBALine())
		}
	}
	return missing
}

func ruleSignature(r models.HBARule) string {
	return strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s|%s|%s|%s", r.RuleType, r.Database, r.DBUser, r.Address)))
}

func startsWithRuleType(s string) bool {
	types := []string{"local", "host", "hostssl", "hostnossl", "hostgssenc", "hostnogssenc"}
	s = strings.ToLower(strings.TrimSpace(s))
	for _, t := range types {
		if strings.HasPrefix(s, t+" ") || strings.HasPrefix(s, t+"\t") {
			return true
		}
	}
	return false
}
