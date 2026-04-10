package services

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
	APIKey    string
	FromEmail string
	FromName  string
}

func NewEmailService() *EmailService {
	return &EmailService{
		APIKey:    os.Getenv("SENDGRID_API_KEY"),
		FromEmail: "noreply@smsystem.com",
		FromName:  "SMSystem",
	}
}

func (e *EmailService) Send(toEmail, toName, subject, htmlContent string) error {
	if e.APIKey == "" {
		fmt.Printf("EmailService: SENDGRID_API_KEY not set, skipping email to %s\n", toEmail)
		return nil
	}

	if toEmail == "" {
		return nil
	}

	from := mail.NewEmail(e.FromName, e.FromEmail)
	to := mail.NewEmail(toName, toEmail)
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)

	client := sendgrid.NewSendClient(e.APIKey)
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid returned status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

func (e *EmailService) SendTransferNotification(toEmail, branchName, refNumber, status, fromBranch, toBranch string) error {
	if toEmail == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] Stock Transfer %s", branchName, status)

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #4f46e5;">Stock Transfer Update</h2>
			<p>Hello,</p>
			<p>A stock transfer has been <strong>%s</strong>.</p>
			<table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
				<tr>
					<td style="padding: 8px; border: 1px solid #ddd;"><strong>Reference</strong></td>
					<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px; border: 1px solid #ddd;"><strong>Status</strong></td>
					<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px; border: 1px solid #ddd;"><strong>From Branch</strong></td>
					<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px; border: 1px solid #ddd;"><strong>To Branch</strong></td>
					<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
				</tr>
			</table>
			<p>Please log in to SMSystem to review and take action if needed.</p>
			<p style="color: #666; font-size: 12px; margin-top: 30px;">This is an automated notification from SMSystem.</p>
		</div>
	`, status, refNumber, status, fromBranch, toBranch)

	return e.Send(toEmail, "Branch Manager", subject, html)
}
