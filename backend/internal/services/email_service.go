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

	subject := fmt.Sprintf("%s [%s] Transfer %s - %s", emoji, branchName, refNumber, label)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin: 0; padding: 0; background-color: #f3f4f6; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;">
  <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #4f46e5 0%%, #7c3aed 100%%); border-radius: 12px 12px 0 0; padding: 30px; text-align: center;">
      <h1 style="color: #ffffff; margin: 0; font-size: 22px; font-weight: 600;">%s Stock Transfer Update</h1>
      <p style="color: #c7d2fe; margin: 8px 0 0 0; font-size: 14px;">Reference: %s</p>
    </div>
    <div style="background: #ffffff; padding: 30px; border-radius: 0 0 12px 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
      <div style="text-align: center; margin-bottom: 24px;">
        <span style="display: inline-block; background-color: %s; color: #ffffff; padding: 8px 20px; border-radius: 20px; font-size: 14px; font-weight: 600;">%s %s</span>
      </div>
      <p style="color: #374151; font-size: 15px; line-height: 1.6; margin-bottom: 24px;">%s</p>
      <table style="width: 100%%; border-collapse: collapse; margin: 0 0 24px 0;">
        <tr>
          <td style="padding: 12px 16px; background: #f9fafb; border: 1px solid #e5e7eb; font-weight: 600; color: #374151; width: 40%%;">Reference</td>
          <td style="padding: 12px 16px; border: 1px solid #e5e7eb; color: #111827; font-family: monospace;">%s</td>
        </tr>
        <tr>
          <td style="padding: 12px 16px; background: #f9fafb; border: 1px solid #e5e7eb; font-weight: 600; color: #374151;">From Branch</td>
          <td style="padding: 12px 16px; border: 1px solid #e5e7eb; color: #111827;">%s</td>
        </tr>
        <tr>
          <td style="padding: 12px 16px; background: #f9fafb; border: 1px solid #e5e7eb; font-weight: 600; color: #374151;">To Branch</td>
          <td style="padding: 12px 16px; border: 1px solid #e5e7eb; color: #111827;">%s</td>
        </tr>
        <tr>
          <td style="padding: 12px 16px; background: #f9fafb; border: 1px solid #e5e7eb; font-weight: 600; color: #374151;">Status</td>
          <td style="padding: 12px 16px; border: 1px solid #e5e7eb;"><span style="color: %s; font-weight: 600;">%s</span></td>
        </tr>
      </table>
      <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;">
      <p style="color: #9ca3af; font-size: 12px; text-align: center; margin: 0;">This is an automated notification from SMSystem.</p>
    </div>
  </div>
</body>
</html>`,
		emoji, refNumber, color, emoji, label, actionMsg,
		refNumber, fromBranch, toBranch, color, label,
	)

	err := e.Send(toEmail, "Branch Manager", subject, html)
	if err != nil {
		log.Printf("[EMAIL] FAILED to send transfer notification to %s for %s: %v", toEmail, refNumber, err)
		return nil
	}

	return nil
}
