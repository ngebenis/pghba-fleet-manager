package services

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/pghba-manager/backend/internal/config"
)

type Notifier struct {
	cfg *config.Config
}

func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{cfg: cfg}
}

type TempAccessAlert struct {
	IPAddress     string
	Database      string
	DBUser        string
	InstanceLabel string
	HostName      string
	GrantedBy     string
	ExpiresAt     time.Time
	MinutesLeft   int
	Reason        string
}

func (n *Notifier) SendTempAccessWarning(alerts []TempAccessAlert) {
	if len(alerts) == 0 || !n.cfg.NotifyEnabled {
		return
	}

	subject := fmt.Sprintf("[PGHBA Manager] ⚠️ %d Temporary Access Expiring Soon", len(alerts))
	body := buildEmailBody(alerts)

	if n.cfg.SMTPEnabled && n.cfg.SMTPHost != "" {
		go func() {
			if err := n.sendEmail(subject, body); err != nil {
				log.Printf("[notifier] email failed: %v", err)
			} else {
				log.Printf("[notifier] email sent for %d expiring temp access", len(alerts))
			}
		}()
	}

	if n.cfg.TelegramToken != "" && n.cfg.TelegramChatID != "" {
		go func() {
			msg := buildTelegramMsg(alerts)
			if err := n.sendTelegram(msg); err != nil {
				log.Printf("[notifier] telegram failed: %v", err)
			}
		}()
	}
}

func (n *Notifier) SendTempAccessExpired(alerts []TempAccessAlert) {
	if len(alerts) == 0 || !n.cfg.NotifyEnabled {
		return
	}

	subject := fmt.Sprintf("[PGHBA Manager] 🔒 %d Temporary Access Expired", len(alerts))
	body := buildExpiredEmailBody(alerts)

	if n.cfg.SMTPEnabled && n.cfg.SMTPHost != "" {
		go func() {
			if err := n.sendEmail(subject, body); err != nil {
				log.Printf("[notifier] expired email failed: %v", err)
			}
		}()
	}

	if n.cfg.TelegramToken != "" && n.cfg.TelegramChatID != "" {
		go func() {
			msg := buildExpiredTelegramMsg(alerts)
			n.sendTelegram(msg)
		}()
	}
}

func (n *Notifier) sendEmail(subject, body string) error {
	addr := net.JoinHostPort(n.cfg.SMTPHost, strconv.Itoa(n.cfg.SMTPPort))
	var auth smtp.Auth
	if n.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", n.cfg.SMTPUser, n.cfg.SMTPPass, n.cfg.SMTPHost)
	}

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
		n.cfg.SMTPFrom, n.cfg.SMTPTo, subject, body,
	))

	return smtp.SendMail(addr, auth, n.cfg.SMTPFrom, strings.Split(n.cfg.SMTPTo, ","), msg)
}

func (n *Notifier) sendTelegram(msg string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.cfg.TelegramToken)
	payload := fmt.Sprintf(`{"chat_id":"%s","text":"%s","parse_mode":"HTML"}`,
		n.cfg.TelegramChatID,
		strings.ReplaceAll(msg, `"`, `\"`))
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func buildEmailBody(alerts []TempAccessAlert) string {
	var sb strings.Builder
	sb.WriteString(`<html><body style="font-family:monospace;background:#0d1117;color:#e6edf3;padding:20px">`)
	sb.WriteString(`<h2 style="color:#f0883e">⚠️ Temporary Access Expiring Soon</h2>`)
	sb.WriteString(`<table style="border-collapse:collapse;width:100%">`)
	sb.WriteString(`<tr style="background:#161b22"><th style="padding:8px;text-align:left;border:1px solid #30363d">IP Address</th><th style="padding:8px;border:1px solid #30363d">Instance</th><th style="padding:8px;border:1px solid #30363d">Expires In</th><th style="padding:8px;border:1px solid #30363d">Granted By</th></tr>`)

	for _, a := range alerts {
		color := "#f0883e"
		if a.MinutesLeft <= 5 {
			color = "#ff7b72"
		}
		sb.WriteString(fmt.Sprintf(
			`<tr><td style="padding:8px;border:1px solid #30363d;color:#58a6ff">%s</td><td style="padding:8px;border:1px solid #30363d">%s</td><td style="padding:8px;border:1px solid #30363d;color:%s;font-weight:bold">%d minutes</td><td style="padding:8px;border:1px solid #30363d">%s</td></tr>`,
			a.IPAddress, a.InstanceLabel, color, a.MinutesLeft, a.GrantedBy,
		))
	}
	sb.WriteString(`</table>`)
	sb.WriteString(`<p style="color:#8b949e;margin-top:16px">PGHBA Manager — PostgreSQL HBA Centralized Management</p>`)
	sb.WriteString(`</body></html>`)
	return sb.String()
}

func buildExpiredEmailBody(alerts []TempAccessAlert) string {
	var sb strings.Builder
	sb.WriteString(`<html><body style="font-family:monospace;background:#0d1117;color:#e6edf3;padding:20px">`)
	sb.WriteString(`<h2 style="color:#3fb950">🔒 Temporary Access Expired</h2>`)
	for _, a := range alerts {
		sb.WriteString(fmt.Sprintf(`<p>✅ <strong>%s</strong> → %s (granted by %s) has expired</p>`, a.IPAddress, a.InstanceLabel, a.GrantedBy))
	}
	sb.WriteString(`</body></html>`)
	return sb.String()
}

func buildTelegramMsg(alerts []TempAccessAlert) string {
	var sb strings.Builder
	sb.WriteString("⚠️ <b>PGHBA Manager — Temp Access Expiring</b>\n\n")
	for _, a := range alerts {
		emoji := "🟡"
		if a.MinutesLeft <= 5 {
			emoji = "🔴"
		}
		sb.WriteString(fmt.Sprintf("%s <code>%s</code> → %s\nExpires in <b>%dm</b> | By: %s\n\n",
			emoji, a.IPAddress, a.InstanceLabel, a.MinutesLeft, a.GrantedBy))
	}
	return sb.String()
}

func buildExpiredTelegramMsg(alerts []TempAccessAlert) string {
	var sb strings.Builder
	sb.WriteString("🔒 <b>PGHBA Manager — Temp Access Expired</b>\n\n")
	for _, a := range alerts {
		sb.WriteString(fmt.Sprintf("✅ <code>%s</code> → %s (by %s)\n", a.IPAddress, a.InstanceLabel, a.GrantedBy))
	}
	return sb.String()
}

// SendRaw — send raw telegram message
func (n *Notifier) SendRaw(msg string) {
	if n.cfg.TelegramToken == "" || n.cfg.TelegramChatID == "" { return }
	go func() {
		if err := n.sendTelegram(msg); err != nil {
			log.Printf("[notifier] SendRaw failed: %v", err)
		}
	}()
}
