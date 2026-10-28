package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

type NotificationItem struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`   // "expiring_soon" | "expired" | "granted" | "revoked"
	Level         string    `json:"level"`  // "info" | "warning" | "danger"
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	IPAddress     string    `json:"ip_address"`
	InstanceLabel string    `json:"instance_label"`
	HostName      string    `json:"host_name"`
	ExpiresAt     time.Time `json:"expires_at"`
	GrantedBy     string    `json:"granted_by"`
	Reason        string    `json:"reason"`
	MinutesLeft   int       `json:"minutes_left"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *NotificationHandler) List(c *gin.Context) {
	now := time.Now()

	type Row struct {
		ID            string    `gorm:"column:id"`
		IPAddress     string    `gorm:"column:ip_address"`
		InstanceName  string    `gorm:"column:instance_name"`
		Port          int       `gorm:"column:port"`
		HostName      string    `gorm:"column:host_name"`
		ExpiresAt     time.Time `gorm:"column:expires_at"`
		Username      string    `gorm:"column:username"`
		Reason        string    `gorm:"column:reason"`
		CreatedAt     time.Time `gorm:"column:created_at"`
	}

	var rows []Row
	h.db.Raw(`
		SELECT
			ta.id,
			COALESCE(r.address::text, '') as ip_address,
			COALESCE(i.instance_name, '')  as instance_name,
			COALESCE(i.port, 0)            as port,
			COALESCE(h.name, '')           as host_name,
			ta.expires_at,
			COALESCE(u.username, 'system') as username,
			COALESCE(ta.reason, '')        as reason,
			ta.created_at
		FROM temp_access ta
		LEFT JOIN hba_rules r    ON r.id = ta.rule_id
		LEFT JOIN pg_instances i ON i.id = ta.instance_id
		LEFT JOIN db_hosts h     ON h.id = i.host_id
		LEFT JOIN users u        ON u.id = ta.granted_by_id
		WHERE ta.revoked = false
		  AND ta.expires_at > NOW() - INTERVAL '5 minutes'
		ORDER BY ta.expires_at ASC
	`).Scan(&rows)

	notifications := []NotificationItem{}

	for _, row := range rows {
		diff := row.ExpiresAt.Sub(now)
		minutesLeft := int(diff.Minutes())
		instanceLabel := row.InstanceName + ":" + itoa(row.Port)

		var notifType, level, title, message string

		if diff < 0 {
			// Just expired (within last 5 min)
			notifType = "expired"
			level = "danger"
			title = "Temporary Access Expired"
			message = "Access for " + row.IPAddress + " to " + instanceLabel + " has expired"
		} else if minutesLeft <= 5 {
			notifType = "expiring_soon"
			level = "danger"
			title = "⚠️ Expires in " + itoa(minutesLeft) + " minute(s)"
			message = row.IPAddress + " → " + instanceLabel
		} else if minutesLeft <= 15 {
			notifType = "expiring_soon"
			level = "warning"
			title = "Expires in " + itoa(minutesLeft) + " minutes"
			message = row.IPAddress + " → " + instanceLabel
		} else if minutesLeft <= 60 {
			notifType = "expiring_soon"
			level = "info"
			title = "Expires in " + itoa(minutesLeft) + " minutes"
			message = row.IPAddress + " → " + instanceLabel
		} else {
			notifType = "active"
			level = "info"
			title = "Active temp access"
			message = row.IPAddress + " → " + instanceLabel
		}

		notifications = append(notifications, NotificationItem{
			ID:            row.ID,
			Type:          notifType,
			Level:         level,
			Title:         title,
			Message:       message,
			IPAddress:     row.IPAddress,
			InstanceLabel: instanceLabel,
			HostName:      row.HostName,
			ExpiresAt:     row.ExpiresAt,
			GrantedBy:     row.Username,
			Reason:        row.Reason,
			MinutesLeft:   minutesLeft,
			CreatedAt:     row.CreatedAt,
		})
	}

	// Count warnings
	warnings := 0
	for _, n := range notifications {
		if n.Level == "warning" || n.Level == "danger" {
			warnings++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         len(notifications),
		"warnings":      warnings,
	})
}

func itoa(n int) string {
	if n < 0 { return "0" }
	if n == 0 { return "0" }
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
