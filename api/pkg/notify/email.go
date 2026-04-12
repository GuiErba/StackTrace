package notify

import (
	"fmt"
	"log"
	"time"

	"github.com/resend/resend-go/v3"
)

type EmailNotifier struct {
	client *resend.Client
	from   string
}

func NewEmailNotifier(apiKey, from string) *EmailNotifier {
	return &EmailNotifier{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

func (n *EmailNotifier) SendAlert(to, projectName, title, description string, errorCount int) error {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto;">
			<div style="background: #ef4444; color: white; padding: 20px; border-radius: 8px 8px 0 0;">
				<h1 style="margin: 0; font-size: 20px;">⚠️ Incident Detected</h1>
				<p style="margin: 5px 0 0; opacity: 0.9;">%s</p>
			</div>
			<div style="background: #fef2f2; padding: 20px; border: 1px solid #fecaca; border-top: none; border-radius: 0 0 8px 8px;">
				<h2 style="margin: 0 0 10px; color: #991b1b; font-size: 18px;">%s</h2>
				<p style="color: #7f1d1d; margin: 0 0 15px;">%s</p>
				<table style="width: 100%%; border-collapse: collapse;">
					<tr>
						<td style="padding: 8px 0; color: #991b1b; font-weight: bold;">Error Count</td>
						<td style="padding: 8px 0; color: #7f1d1d;">%d errors in the time window</td>
					</tr>
					<tr>
						<td style="padding: 8px 0; color: #991b1b; font-weight: bold;">Detected At</td>
						<td style="padding: 8px 0; color: #7f1d1d;">%s</td>
					</tr>
				</table>
			</div>
			<p style="color: #9ca3af; font-size: 12px; margin-top: 15px;">Sent by StackTrace</p>
		</div>
	`, projectName, title, description, errorCount, time.Now().UTC().Format(time.RFC3339))

	params := &resend.SendEmailRequest{
		From:    n.from,
		To:      []string{to},
		Subject: fmt.Sprintf("[StackTrace] Incident: %s — %s", title, projectName),
		Html:    html,
	}

	_, err := n.client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send alert email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Alert email sent to %s for project %s", to, projectName)
	return nil
}
