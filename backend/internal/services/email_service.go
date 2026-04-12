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
		APIKey:    os.Getenv("BREVO_API_KEY"),
		FromEmail: getEmailEnv("BREVO_FROM_EMAIL", "johndinglasan12@gmail.com"),
		FromName:  getEmailEnv("BREVO_FROM_NAME", "SMSystem"),
		AdminBcc:  os.Getenv("ADMIN_BCC_EMAIL"),
	}
}

func getEmailEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoPayload struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Bcc         []brevoContact `json:"bcc,omitempty"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

func (e *EmailService) Send(toEmail, toName, subject, htmlContent string) error {
	if e.APIKey == "" {
		log.Printf("[EMAIL] BREVO_API_KEY not set, skipping email to %s | Subject: %s", toEmail, subject)
		return nil
	}

	if toEmail == "" {
		return nil
	}

	payload := brevoPayload{
		Sender:      brevoContact{Email: e.FromEmail, Name: e.FromName},
		To:          []brevoContact{{Email: toEmail, Name: toName}},
		Subject:     subject,
		HTMLContent: htmlContent,
	}

	if e.AdminBcc != "" && e.AdminBcc != toEmail {
		payload.Bcc = []brevoContact{{Email: e.AdminBcc}}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", e.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[EMAIL] ERROR sending to %s: %v", toEmail, err)
		return fmt.Errorf("brevo API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("[EMAIL] ERROR Brevo returned %d to %s: %s", resp.StatusCode, toEmail, string(respBody))
		return fmt.Errorf("brevo returned status %d: %s", resp.StatusCode, string(respBody))
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

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#ffffff;">
  <table width="100%%" cellpadding="0" cellspacing="0">
    <tr><td align="center" style="padding:40px 20px;">
      <table width="480" cellpadding="0" cellspacing="0" style="max-width:480px;">

        <tr><td style="border-top:4px solid %s;padding-top:32px;">
          <p style="margin:0 0 4px;font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:2px;font-weight:600;">SMSystem</p>
          <h1 style="margin:0 0 24px;font-size:22px;color:#111827;font-weight:700;">%s %s</h1>
        </td></tr>

        <tr><td style="padding-bottom:24px;">
          <p style="margin:0;font-size:15px;color:#374151;line-height:1.6;">%s</p>
        </td></tr>

        <tr><td style="padding:20px;background:#f9fafb;border-radius:8px;">
          <table width="100%%" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding:6px 0;color:#6b7280;font-size:13px;width:100px;">Reference</td>
              <td style="padding:6px 0;color:#111827;font-size:13px;font-weight:600;font-family:monospace;">%s</td>
            </tr>
            <tr>
              <td style="padding:6px 0;color:#6b7280;font-size:13px;">From</td>
              <td style="padding:6px 0;color:#111827;font-size:13px;">%s</td>
            </tr>
            <tr>
              <td style="padding:6px 0;color:#6b7280;font-size:13px;">To</td>
              <td style="padding:6px 0;color:#111827;font-size:13px;">%s</td>
            </tr>
            <tr>
              <td style="padding:6px 0;color:#6b7280;font-size:13px;">Status</td>
              <td style="padding:6px 0;"><span style="display:inline-block;background:%s;color:#fff;padding:2px 10px;border-radius:10px;font-size:12px;font-weight:600;">%s</span></td>
            </tr>
          </table>
        </td></tr>

        <tr><td style="padding-top:32px;border-top:1px solid #e5e7eb;margin-top:24px;">
          <p style="margin:0;font-size:11px;color:#9ca3af;">This is an automated notification from SMSystem.</p>
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`,
		color, emoji, label,
		actionMsg,
		refNumber, fromBranch, toBranch, color, label,
	)

	err := e.Send(toEmail, "Branch Manager", subject, html)
	if err != nil {
		log.Printf("[EMAIL] FAILED to send transfer notification to %s for %s: %v", toEmail, refNumber, err)
		return nil
	}

	return nil
}

// SendPromoEmail sends a promotional email for tire sales
func (e *EmailService) SendPromoEmail(toEmail, toName, promoCode, discount string) error {
	if e.APIKey == "" {
		log.Printf("[EMAIL] BREVO_API_KEY not set, skipping promo email to %s", toEmail)
		return nil
	}

	if toEmail == "" {
		return nil
	}

	subject := "🔥 Limited Time Offer! Get Premium Tires at Special Prices"

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Special Tire Offer</title>
</head>
<body style="margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background-color:#1a1a2e;">
  <table width="100%%" cellpadding="0" cellspacing="0">
    <tr>
      <td align="center" style="padding:40px 20px;">
        <table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;background:#ffffff;border-radius:16px;overflow:hidden;">
          <!-- Header with Logo -->
          <tr>
            <td style="background:linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);padding:30px;text-align:center;">
              <h1 style="color:#ffffff;font-size:28px;margin:0;font-weight:800;">🛞 SMSYSTEM</h1>
              <p style="color:rgba(255,255,255,0.9);margin:10px 0 0;font-size:14px;">Premium Tire Solutions</p>
            </td>
          </tr>
          
          <!-- Hero Banner -->
          <tr>
            <td style="padding:0;position:relative;">
              <div style="background:linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%);padding:40px 30px;text-align:center;">
                <p style="color:#ffffff;font-size:14px;font-weight:600;margin:0;text-transform:uppercase;letter-spacing:2px;">Special Offer</p>
                <h2 style="color:#ffffff;font-size:36px;margin:15px 0;font-weight:800;">BUY 4 GET 1 FREE!</h2>
                <p style="color:rgba(255,255,255,0.9);font-size:16px;margin:0;">On all premium tire brands</p>
              </div>
            </td>
          </tr>
          
          <!-- Main Content -->
          <tr>
            <td style="padding:40px 30px;">
              <p style="color:#333333;font-size:16px;margin:0;line-height:1.6;">
                Hello <strong>%s</strong>! 👋
              </p>
              <p style="color:#666666;font-size:15px;margin:20px 0;line-height:1.6;">
                We have an exclusive offer just for you! Get premium quality tires at unbeatable prices. Whether you need tires for your car, truck, or SUV - we've got you covered with top brands!
              </p>
              
              <!-- Features -->
              <table width="100%%" cellpadding="0" cellspacing="0" style="margin:25px 0;">
                <tr>
                  <td style="text-align:center;padding:15px;">
                    <div style="width:60px;height:60px;background:#fff3e0;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;margin-bottom:10px;">🚚</div>
                    <p style="color:#333;font-size:13px;font-weight:600;margin:0;">Free Delivery</p>
                  </td>
                  <td style="text-align:center;padding:15px;">
                    <div style="width:60px;height:60px;background:#e8f5e9;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;margin-bottom:10px;">✓</div>
                    <p style="color:#333;font-size:13px;font-weight:600;margin:0;">Quality Guaranteed</p>
                  </td>
                  <td style="text-align:center;padding:15px;">
                    <div style="width:60px;height:60px;background:#e3f2fd;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;margin-bottom:10px;">🔧</div>
                    <p style="color:#333;font-size:13px;font-weight:600;margin:0;">Free Installation</p>
                  </td>
                </tr>
              </table>
              
              <!-- Discount Code Box -->
              <div style="background:#f8f9fa;border:2px dashed #667eea;border-radius:12px;padding:25px;text-align:center;margin:30px 0;">
                <p style="color:#667eea;font-size:14px;font-weight:600;margin:0;text-transform:uppercase;letter-spacing:1px;">Use Code</p>
                <p style="color:#333;font-size:32px;font-weight:800;margin:10px 0;font-family:monospace;letter-spacing:4px;">%s</p>
                <p style="color:#999;font-size:12px;margin:0;">Valid until end of this month</p>
              </div>
              
              <!-- CTA Button -->
              <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center">
                    <a href="https://smstyredepot.com" style="display:inline-block;background:linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);color:#ffffff;font-size:16px;font-weight:700;padding:16px 40px;border-radius:50px;text-decoration:none;text-transform:uppercase;letter-spacing:1px;">
                      Shop Now →
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Tire Categories -->
          <tr>
            <td style="background:#f8f9fa;padding:30px;">
              <p style="color:#333;font-size:14px;font-weight:700;margin:0 0 20px;text-align:center;text-transform:uppercase;letter-spacing:1px;">Popular Categories</p>
              <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center" style="padding:10px;">
                    <div style="background:#fff;border-radius:12px;padding:15px;width:80px;">
                      <p style="font-size:30px;margin:0;">🚗</p>
                      <p style="color:#666;font-size:11px;margin:5px 0 0;">Car Tires</p>
                    </div>
                  </td>
                  <td align="center" style="padding:10px;">
                    <div style="background:#fff;border-radius:12px;padding:15px;width:80px;">
                      <p style="font-size:30px;margin:0;">🚚</p>
                      <p style="color:#666;font-size:11px;margin:5px 0 0;">Truck Tires</p>
                    </div>
                  </td>
                  <td align="center" style="padding:10px;">
                    <div style="background:#fff;border-radius:12px;padding:15px;width:80px;">
                      <p style="font-size:30px;margin:0;">🏎️</p>
                      <p style="color:#666;font-size:11px;margin:5px 0 0;">Sports</p>
                    </div>
                  </td>
                  <td align="center" style="padding:10px;">
                    <div style="background:#fff;border-radius:12px;padding:15px;width:80px;">
                      <p style="font-size:30px;margin:0;">⛺</p>
                      <p style="color:#666;font-size:11px;margin:5px 0 0;">AT/MT</p>
                    </div>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Footer -->
          <tr>
            <td style="background:#1a1a2e;padding:30px;text-align:center;">
              <p style="color:rgba(255,255,255,0.7);font-size:13px;margin:0;">
                📧 info@smstyredepot.com | 📞 +63 911-111-1111
              </p>
              <p style="color:rgba(255,255,255,0.5);font-size:11px;margin:15px 0 0;">
                © 2026 SMSystem. All rights reserved.<br>
                Lipa City, Batangas, Philippines
              </p>
            </td>
          </tr>
        </table>
        
        <p style="color:rgba(255,255,255,0.4);font-size:11px;margin:20px 0 0;text-align:center;">
          This email was sent to %s because you're a valued customer of SMSystem.
        </p>
      </td>
    </tr>
  </table>
</body>
</html>`, toName, promoCode, toEmail)

	err := e.Send(toEmail, toName, subject, html)
	if err != nil {
		log.Printf("[EMAIL] FAILED to send promo email to %s: %v", toEmail, err)
		return err
	}

	log.Printf("[EMAIL] Promo email sent to %s with code %s", toEmail, promoCode)
	return nil
}
