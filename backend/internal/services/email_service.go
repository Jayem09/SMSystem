package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type EmailService struct {
	APIKey    string
	FromEmail string
	FromName  string
	AdminBcc  string
}

func NewEmailService() *EmailService {
	return &EmailService{
		APIKey:    os.Getenv("RESEND_API_KEY"),
		FromEmail: getEmailEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
		FromName:  getEmailEnv("RESEND_FROM_NAME", "SMSystem"),
		AdminBcc:  os.Getenv("ADMIN_BCC_EMAIL"),
	}
}

func getEmailEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Bcc     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (e *EmailService) Send(toEmail, toName, subject, htmlContent string) error {
	if e.APIKey == "" {
		log.Printf("[EMAIL] RESEND_API_KEY not set, skipping email to %s | Subject: %s", toEmail, subject)
		return nil
	}

	if toEmail == "" {
		return nil
	}

	payload := resendPayload{
		From:    fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail),
		To:      []string{toEmail},
		Subject: subject,
		HTML:    htmlContent,
	}

	if e.AdminBcc != "" && e.AdminBcc != toEmail {
		payload.Bcc = []string{e.AdminBcc}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+e.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[EMAIL] ERROR sending to %s: %v", toEmail, err)
		return fmt.Errorf("resend API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("[EMAIL] ERROR Resend returned %d to %s: %s", resp.StatusCode, toEmail, string(respBody))
		return fmt.Errorf("resend returned status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[EMAIL] OK sent to %s | Subject: %s", toEmail, subject)
	return nil
}

// Status display helpers
func statusLabel(status string) string {
	labels := map[string]string{
		"pending":    "Pending Approval",
		"approved":   "Approved",
		"in_transit": "Shipped / In Transit",
		"completed":  "Received & Completed",
		"rejected":   "Rejected",
		"cancelled":  "Cancelled",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return strings.Title(status)
}

func statusColor(status string) string {
	colors := map[string]string{
		"pending":    "#f59e0b",
		"approved":   "#3b82f6",
		"in_transit": "#8b5cf6",
		"completed":  "#10b981",
		"rejected":   "#ef4444",
		"cancelled":  "#6b7280",
	}
	if color, ok := colors[status]; ok {
		return color
	}
	return "#4f46e5"
}

// adjustColor returns a second gradient color for the header
func adjustColor(hex string) string {
	gradients := map[string]string{
		"#f59e0b": "#d97706",
		"#3b82f6": "#2563eb",
		"#8b5cf6": "#7c3aed",
		"#10b981": "#059669",
		"#ef4444": "#dc2626",
		"#6b7280": "#4b5563",
		"#4f46e5": "#4338ca",
	}
	if g, ok := gradients[hex]; ok {
		return g
	}
	return hex
}

func statusEmoji(status string) string {
	emojis := map[string]string{
		"pending":    "⏳",
		"approved":   "✅",
		"in_transit": "🚚",
		"completed":  "📦",
		"rejected":   "❌",
		"cancelled":  "🚫",
	}
	if emoji, ok := emojis[status]; ok {
		return emoji
	}
	return "📋"
}

func statusActionMessage(status, recipientType string) string {
	if recipientType == "source" {
		switch status {
		case "approved":
			return "Your transfer request has been approved. Please prepare the items for shipment and click <strong>Ship</strong> when ready."
		case "in_transit":
			return "The items have been shipped and are now in transit to the destination branch."
		case "completed":
			return "The destination branch has confirmed receipt of all items. This transfer is now complete."
		case "rejected":
			return "This transfer request has been rejected. Please review and create a new request if needed."
		case "cancelled":
			return "This transfer has been cancelled. No stock was moved."
		default:
			return "The transfer status has been updated."
		}
	}
	switch status {
	case "in_transit":
		return "A shipment is on its way to your branch. Please prepare to receive and verify the items, then click <strong>Receive</strong> to confirm."
	case "completed":
		return "You have confirmed receipt of all items. This transfer is now complete."
	default:
		return "The transfer status has been updated."
	}
}

func (e *EmailService) SendTransferNotification(toEmail, branchName, refNumber, status, fromBranch, toBranch string) error {
	if toEmail == "" {
		log.Printf("[EMAIL] No email address for branch %s, skipping notification", branchName)
		return nil
	}

	log.Printf("[EMAIL] Sending transfer notification to %s (ref: %s, status: %s)", toEmail, refNumber, status)

	recipientType := "source"
	if branchName == toBranch {
		recipientType = "destination"
	}

	emoji := statusEmoji(status)
	label := statusLabel(status)
	color := statusColor(status)
	actionMsg := statusActionMessage(status, recipientType)

	subject := fmt.Sprintf("%s [%s] Transfer %s — %s", emoji, branchName, refNumber, label)

	// Progress tracker dots
	progressDots := func() string {
		steps := []struct{ name, status string }{
			{"Pending", "pending"},
			{"Approved", "approved"},
			{"Shipped", "in_transit"},
			{"Received", "completed"},
		}
		dots := ""
		reached := false
		for i, step := range steps {
			isActive := step.status == status
			isPast := !reached && !isActive
			dotColor := "#d1d5db"
			textColor := "#9ca3af"
			if isActive {
				dotColor = color
				textColor = color
				reached = true
			} else if isPast {
				dotColor = "#10b981"
				textColor = "#10b981"
			}
			if isActive {
				reached = true
			}
			lineColor := "#e5e7eb"
			if isPast {
				lineColor = "#10b981"
			}
			connector := ""
			if i > 0 {
				connector = fmt.Sprintf(`<td style="padding:0;"><div style="width:40px;height:3px;background:%s;"></div></td>`, lineColor)
			}
			dots += fmt.Sprintf(`%s<td style="padding:0;text-align:center;"><div style="width:14px;height:14px;border-radius:50%%;background:%s;margin:0 auto;"></div><div style="font-size:10px;color:%s;margin-top:4px;white-space:nowrap;">%s</div></td>`, connector, dotColor, textColor, step.name)
		}
		return dots
	}()

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#0f172a;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <div style="max-width:560px;margin:0 auto;padding:32px 16px;">

    <div style="text-align:center;margin-bottom:24px;">
      <div style="display:inline-block;background:linear-gradient(135deg,#6366f1,#8b5cf6);padding:10px 24px;border-radius:12px;">
        <span style="color:#fff;font-size:18px;font-weight:700;letter-spacing:0.5px;">SM</span><span style="color:#c4b5fd;font-size:18px;font-weight:300;">System</span>
      </div>
    </div>

    <div style="background:#1e293b;border-radius:16px;overflow:hidden;border:1px solid #334155;">

      <div style="background:linear-gradient(135deg,%s 0%%,%s 100%%);padding:28px 32px;text-align:center;">
        <div style="font-size:36px;margin-bottom:8px;">%s</div>
        <h1 style="color:#fff;margin:0;font-size:20px;font-weight:700;letter-spacing:0.3px;">%s</h1>
        <p style="color:rgba(255,255,255,0.7);margin:6px 0 0;font-size:13px;">Transfer %s</p>
      </div>

      <div style="padding:28px 32px;">

        <table style="width:100%%;margin:0 auto 28px;"><tr style="vertical-align:top;">%s</tr></table>

        <div style="background:#0f172a;border-radius:12px;padding:20px;margin-bottom:24px;">
          <p style="color:#e2e8f0;font-size:14px;line-height:1.7;margin:0;">%s</p>
        </div>

        <table style="width:100%%;border-collapse:separate;border-spacing:0;margin-bottom:24px;">
          <tr>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.08);border:1px solid #334155;border-radius:10px 0 0 0;font-weight:600;color:#94a3b8;font-size:12px;text-transform:uppercase;letter-spacing:1px;width:38%%;">Reference</td>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.04);border:1px solid #334155;border-left:none;border-radius:0 10px 0 0;color:#f1f5f9;font-family:'SF Mono',monospace;font-size:14px;font-weight:600;">%s</td>
          </tr>
          <tr>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.08);border:1px solid #334155;border-top:none;font-weight:600;color:#94a3b8;font-size:12px;text-transform:uppercase;letter-spacing:1px;">From</td>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.04);border:1px solid #334155;border-left:none;border-top:none;color:#f1f5f9;font-size:14px;">📤 %s</td>
          </tr>
          <tr>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.08);border:1px solid #334155;border-top:none;font-weight:600;color:#94a3b8;font-size:12px;text-transform:uppercase;letter-spacing:1px;">To</td>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.04);border:1px solid #334155;border-left:none;border-top:none;color:#f1f5f9;font-size:14px;">📥 %s</td>
          </tr>
          <tr>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.08);border:1px solid #334155;border-top:none;border-radius:0 0 0 10px;font-weight:600;color:#94a3b8;font-size:12px;text-transform:uppercase;letter-spacing:1px;">Status</td>
            <td style="padding:14px 16px;background:rgba(99,102,241,0.04);border:1px solid #334155;border-left:none;border-top:none;border-radius:0 0 10px 0;"><span style="color:%s;font-weight:700;font-size:14px;">%s %s</span></td>
          </tr>
        </table>
      </div>
    </div>

    <div style="text-align:center;margin-top:20px;">
      <p style="color:#475569;font-size:11px;margin:0;">Automated notification from SMSystem</p>
      <p style="color:#334155;font-size:10px;margin:4px 0 0;">Do not reply to this email</p>
    </div>
  </div>
</body>
</html>`,
		color, adjustColor(color), emoji, label, refNumber,
		progressDots,
		actionMsg,
		refNumber, fromBranch, toBranch, color, emoji, label,
	)

	err := e.Send(toEmail, "Branch Manager", subject, html)
	if err != nil {
		log.Printf("[EMAIL] FAILED to send transfer notification to %s for %s: %v", toEmail, refNumber, err)
		return nil
	}

	return nil
}
