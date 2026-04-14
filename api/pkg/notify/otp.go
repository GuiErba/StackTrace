package notify

import (
	"fmt"
	"log"

	"github.com/resend/resend-go/v3"
)

func (n *EmailNotifier) SendOTP(to, code string) error {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto;">
			<div style="background: #0f172a; color: white; padding: 30px; border-radius: 12px;">
				<h1 style="margin: 0 0 8px; font-size: 22px; font-weight: 600;">StackTrace</h1>
				<p style="margin: 0 0 24px; color: #94a3b8; font-size: 14px;">Your verification code</p>
				<div style="background: #1e293b; border: 1px solid #334155; border-radius: 8px; padding: 20px; text-align: center; margin-bottom: 24px;">
					<span style="font-size: 36px; font-weight: 700; letter-spacing: 8px; color: #f8fafc;">%s</span>
				</div>
				<p style="color: #94a3b8; font-size: 13px; margin: 0;">This code expires in 10 minutes. If you didn't request this, ignore this email.</p>
			</div>
		</div>
	`, code)

	params := &resend.SendEmailRequest{
		From:    n.from,
		To:      []string{to},
		Subject: fmt.Sprintf("StackTrace — Your code is %s", code),
		Html:    html,
	}

	_, err := n.client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send OTP email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("OTP email sent to %s", to)
	return nil
}
