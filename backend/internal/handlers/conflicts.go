package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type ConflictHandler struct {
	db *gorm.DB
}

func NewConflictHandler(db *gorm.DB) *ConflictHandler {
	return &ConflictHandler{db: db}
}

type ConflictResult struct {
	Type          string          `json:"type"`
	Severity      string          `json:"severity"`
	Description   string          `json:"description"`
	RuleA         *models.HBARule `json:"rule_a"`
	RuleB         *models.HBARule `json:"rule_b"`
	InstanceID    string          `json:"instance_id"`
	InstanceLabel string          `json:"instance_label"`
	Suggestion    string          `json:"suggestion"`
	OrderA        int             `json:"order_a"`
	OrderB        int             `json:"order_b"`
}

func (h *ConflictHandler) Detect(c *gin.Context) {
	instanceID := c.Query("instance_id")

	var rules []models.HBARule
	q := h.db.Preload("Instance").Preload("Instance.Host").
		Where("enabled = true AND is_temporary = false")
	if instanceID != "" {
		q = q.Where("instance_id = ?", instanceID)
	}
	q.Order("instance_id, rule_order ASC").Find(&rules)

	var conflicts []ConflictResult

	// Group by instance
	byInstance := map[string][]models.HBARule{}
	for _, r := range rules {
		id := r.InstanceID.String()
		byInstance[id] = append(byInstance[id], r)
	}

	for instID, instRules := range byInstance {
		conflicts = append(conflicts, detectConflicts(instID, instRules)...)
	}

	// Stats
	errors   := 0
	warnings := 0
	for _, c := range conflicts {
		if c.Severity == "error"   { errors++ }
		if c.Severity == "warning" { warnings++ }
	}

	c.JSON(http.StatusOK, gin.H{
		"conflicts": conflicts,
		"total":     len(conflicts),
		"errors":    errors,
		"warnings":  warnings,
		"instances_checked": len(byInstance),
	})
}

func detectConflicts(instanceID string, rules []models.HBARule) []ConflictResult {
	var results []ConflictResult

	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			a, b := rules[i], rules[j]

			// 1. Exact duplicate
			if a.RuleType == b.RuleType &&
				a.Database == b.Database &&
				a.DBUser == b.DBUser &&
				a.Address == b.Address &&
				a.AuthMethod == b.AuthMethod {
				results = append(results, ConflictResult{
					Type:          "duplicate",
					Severity:      "error",
					Description:   "Exact duplicate rule detected",
					RuleA:         &rules[i],
					RuleB:         &rules[j],
					InstanceID:    instanceID,
					InstanceLabel: instLabel(rules[i]),
					Suggestion:    "Delete one of the duplicate rules",
					OrderA:        rules[i].RuleOrder,
					OrderB:        rules[j].RuleOrder,
				})
				continue
			}

			// 2. Same address/db/user but different method (shadow)
			if a.RuleType == b.RuleType &&
				a.Database == b.Database &&
				a.DBUser == b.DBUser &&
				a.Address == b.Address &&
				a.AuthMethod != b.AuthMethod {
				// Rule that comes first wins
				results = append(results, ConflictResult{
					Type:        "shadow",
					Severity:    "warning",
					Description: "Same IP/DB/User but different auth method — rule with lower order takes precedence",
					RuleA:       &rules[i],
					RuleB:       &rules[j],
					InstanceID:    instanceID,
					InstanceLabel: instLabel(rules[i]),
					OrderA:        rules[i].RuleOrder,
					OrderB:        rules[j].RuleOrder,
					Suggestion:  "Remove the shadowed rule or change auth method to match",
				})
				continue
			}

			// 3. Broad rule shadows specific (e.g., all/all/10.0.0.0/8 before all/all/10.1.3.0/24)
			if a.RuleType == b.RuleType &&
				a.Database == b.Database &&
				a.DBUser == b.DBUser &&
				i < j && // a comes before b in order
				isBroaderCIDR(a.Address, b.Address) {
				results = append(results, ConflictResult{
					Type:        "shadow",
					Severity:    "warning",
					Description: "Broad rule may shadow more specific rule below it",
					RuleA:       &rules[i],
					RuleB:       &rules[j],
					InstanceID:    instanceID,
					InstanceLabel: instLabel(rules[i]),
					OrderA:        rules[i].RuleOrder,
					OrderB:        rules[j].RuleOrder,
					Suggestion:  "Move the more specific rule above the broad rule",
				})
			}

			// 4. Trust before md5 for same subnet
			if a.RuleType == b.RuleType &&
				a.Database == b.Database &&
				a.DBUser == b.DBUser &&
				a.AuthMethod == "trust" &&
				b.AuthMethod == "md5" &&
				i < j &&
				addressOverlaps(a.Address, b.Address) {
				results = append(results, ConflictResult{
					Type:        "security",
					Severity:    "error",
					Description: "SECURITY: 'trust' rule appears before 'md5' rule for overlapping addresses",
					RuleA:       &rules[i],
					RuleB:       &rules[j],
					InstanceID:    instanceID,
					InstanceLabel: instLabel(rules[i]),
					OrderA:        rules[i].RuleOrder,
					OrderB:        rules[j].RuleOrder,
					Suggestion:  "Move trust rule after md5, or restrict trust to localhost only",
				})
			}
		}
	}
	return results
}

// isBroaderCIDR returns true if cidrA is broader than cidrB
func isBroaderCIDR(cidrA, cidrB string) bool {
	if cidrA == "" || cidrB == "" {
		return false
	}
	partsA := strings.Split(cidrA, "/")
	partsB := strings.Split(cidrB, "/")
	if len(partsA) != 2 || len(partsB) != 2 {
		return false
	}
	// Simpler check: same prefix family, A has shorter prefix = broader
	prefixA := partsA[1]
	prefixB := partsB[1]
	// Convert to int comparison
	a, b := 0, 0
	for _, c := range prefixA { a = a*10 + int(c-'0') }
	for _, c := range prefixB { b = b*10 + int(c-'0') }
	return a < b && strings.HasPrefix(partsB[0], partsA[0][:min3(len(partsA[0]), 7)])
}

// addressOverlaps checks if two CIDR addresses could overlap
func addressOverlaps(a, b string) bool {
	if a == b { return true }
	partsA := strings.Split(a, "/")
	partsB := strings.Split(b, "/")
	if len(partsA) < 1 || len(partsB) < 1 { return false }
	// Simple: check if one is prefix of other
	ipA := strings.Split(partsA[0], ".")
	ipB := strings.Split(partsB[0], ".")
	if len(ipA) >= 3 && len(ipB) >= 3 {
		return ipA[0] == ipB[0] && ipA[1] == ipB[1] && ipA[2] == ipB[2]
	}
	return false
}

func instLabel(r models.HBARule) string {
	if r.Instance.InstanceName != "" {
		return r.Instance.InstanceName + ":" + itoa2(r.Instance.Port)
	}
	return ""
}

func itoa2(n int) string {
	if n == 0 { return "0" }
	result := ""
	for n > 0 { result = string(rune('0'+n%10)) + result; n /= 10 }
	return result
}



func min3(a, b int) int {
	if a < b { return a }
	return b
}
