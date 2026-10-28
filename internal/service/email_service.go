package service

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/quocbaong/profile-backend/internal/config"
	"github.com/quocbaong/profile-backend/internal/models"
)

type EmailService interface {
	SendAdminNotification(msg *models.ContactMessage) error
	SendAutoReply(msg *models.ContactMessage) error
	IsConfigured() bool
}

type smtpEmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &smtpEmailService{cfg: cfg}
}

func (s *smtpEmailService) IsConfigured() bool {
	return s.cfg.ResendAPIKey != "" || s.cfg.BrevoAPIKey != "" ||
		(s.cfg.SMTPHost != "" && s.cfg.SMTPPort != "" && s.cfg.SMTPUser != "" && s.cfg.SMTPPassword != "")
}

// SendAdminNotification sends an alert email to the admin/portfolio owner
func (s *smtpEmailService) SendAdminNotification(msg *models.ContactMessage) error {
	if !s.IsConfigured() {
		log.Println("[EmailService] SMTP not configured; skipping admin notification email")
		return nil
	}

	adminEmail := s.cfg.AdminEmail
	if adminEmail == "" {
		adminEmail = s.cfg.SMTPUser
	}
	if adminEmail == "" {
		return nil
	}

	fullName := strings.TrimSpace(msg.Name + " " + msg.LastName)
	subject := fmt.Sprintf("📬 Tin nhắn liên hệ mới từ %s", fullName)
	if msg.Subject != "" {
		subject = fmt.Sprintf("📬 [%s] Liên hệ từ %s", msg.Subject, fullName)
	}

	body := s.buildAdminNotificationHTML(msg, fullName)
	return s.sendMail([]string{adminEmail}, msg.Email, subject, body)
}

// SendAutoReply sends a polite confirmation email to the visitor
func (s *smtpEmailService) SendAutoReply(msg *models.ContactMessage) error {
	if !s.IsConfigured() {
		log.Println("[EmailService] SMTP not configured; skipping auto-reply email")
		return nil
	}

	if msg.Email == "" {
		return nil
	}

	fullName := strings.TrimSpace(msg.Name + " " + msg.LastName)
	if fullName == "" {
		fullName = "bạn"
	}

	subject := "✨ Cảm ơn bạn đã liên hệ - Nguyễn Quốc Bảo"
	body := s.buildAutoReplyHTML(msg, fullName)
	adminEmail := s.cfg.AdminEmail
	if adminEmail == "" {
		adminEmail = s.cfg.SMTPUser
	}
	return s.sendMail([]string{msg.Email}, adminEmail, subject, body)
}

// sendMail delegates to Resend API (443), Brevo API (443), or standard SMTP (587/465)
func (s *smtpEmailService) sendMail(to []string, replyTo string, subject string, htmlBody string) error {
	if s.cfg.ResendAPIKey != "" {
		return s.sendViaResend(to, replyTo, subject, htmlBody)
	}
	if s.cfg.BrevoAPIKey != "" {
		return s.sendViaBrevo(to, replyTo, subject, htmlBody)
	}
	return s.sendViaSMTP(to, replyTo, subject, htmlBody)
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	ReplyTo string   `json:"reply_to,omitempty"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *smtpEmailService) sendViaResend(to []string, replyTo string, subject string, htmlBody string) error {
	from := s.cfg.SMTPFrom
	if from == "" {
		from = "Portfolio <onboarding@resend.dev>"
	}

	reqBody := resendRequest{
		From:    from,
		To:      to,
		ReplyTo: replyTo,
		Subject: subject,
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("resend: failed to marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("resend: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(s.cfg.ResendAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: request failed over https 443: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend: api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("[EmailService] Successfully sent email to %v via Resend API (HTTPS port 443)", to)
	return nil
}

type brevoSender struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type brevoRecipient struct {
	Email string `json:"email"`
}

type brevoReplyTo struct {
	Email string `json:"email"`
}

type brevoRequest struct {
	Sender      brevoSender      `json:"sender"`
	To          []brevoRecipient `json:"to"`
	ReplyTo     *brevoReplyTo    `json:"replyTo,omitempty"`
	Subject     string           `json:"subject"`
	HTMLContent string           `json:"htmlContent"`
}

func (s *smtpEmailService) sendViaBrevo(to []string, replyTo string, subject string, htmlBody string) error {
	senderEmail := s.cfg.SMTPUser
	if senderEmail == "" {
		senderEmail = "baonq8828@gmail.com"
	}
	senderName := "Nguyễn Quốc Bảo"

	var recipients []brevoRecipient
	for _, t := range to {
		recipients = append(recipients, brevoRecipient{Email: t})
	}

	reqBody := brevoRequest{
		Sender:      brevoSender{Name: senderName, Email: senderEmail},
		To:          recipients,
		Subject:     subject,
		HTMLContent: htmlBody,
	}

	if replyTo != "" {
		reqBody.ReplyTo = &brevoReplyTo{Email: replyTo}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("brevo: failed to marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("brevo: failed to create request: %w", err)
	}

	req.Header.Set("api-key", strings.TrimSpace(s.cfg.BrevoAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo: request failed over https 443: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("brevo: api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("[EmailService] Successfully sent email to %v via Brevo API (HTTPS port 443)", to)
	return nil
}

// sendViaSMTP delivers the email via STARTTLS (port 587/25) or direct TLS (port 465)
func (s *smtpEmailService) sendViaSMTP(to []string, replyTo string, subject string, htmlBody string) error {
	host := s.cfg.SMTPHost
	port := s.cfg.SMTPPort
	addr := net.JoinHostPort(host, port)

	fromHeader := s.cfg.SMTPFrom
	if fromHeader == "" {
		fromHeader = fmt.Sprintf("Nguyễn Quốc Bảo <%s>", s.cfg.SMTPUser)
	}

	// Envelope sender is always the authenticated user for strict SPF/DKIM compliance
	fromAddr := s.cfg.SMTPUser

	// Encode subject with UTF-8 Base64 to prevent header corruption
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	dateHeader := time.Now().Format(time.RFC1123Z)
	toHeader := strings.Join(to, ", ")

	replyToLine := ""
	if replyTo != "" {
		replyToLine = fmt.Sprintf("Reply-To: %s\r\n", replyTo)
	}

	rawMessage := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"%s"+
		"Subject: %s\r\n"+
		"Date: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"Content-Transfer-Encoding: base64\r\n\r\n"+
		"%s\r\n",
		fromHeader,
		toHeader,
		replyToLine,
		encodedSubject,
		dateHeader,
		base64.StdEncoding.EncodeToString([]byte(htmlBody)),
	)

	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, host)

	// Direct SSL/TLS (typically port 465)
	if port == "465" {
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial failed on %s: %w", addr, err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client creation failed: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}

		if err = client.Mail(fromAddr); err != nil {
			return fmt.Errorf("smtp mail from failed: %w", err)
		}

		for _, recipient := range to {
			if err = client.Rcpt(recipient); err != nil {
				return fmt.Errorf("smtp rcpt failed for %s: %w", recipient, err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data writer failed: %w", err)
		}
		if _, err = w.Write([]byte(rawMessage)); err != nil {
			_ = w.Close()
			return fmt.Errorf("smtp write message failed: %w", err)
		}
		if err = w.Close(); err != nil {
			return fmt.Errorf("smtp close writer failed: %w", err)
		}
		return nil
	}

	// STARTTLS (typically port 587 or 25)
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial failed on %s: %w", addr, err)
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err = client.Mail(fromAddr); err != nil {
		return fmt.Errorf("smtp mail from failed: %w", err)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp rcpt failed for %s: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data writer failed: %w", err)
	}
	if _, err = w.Write([]byte(rawMessage)); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp write message failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close writer failed: %w", err)
	}

	return nil
}

func (s *smtpEmailService) buildAdminNotificationHTML(msg *models.ContactMessage, fullName string) string {
	createdTime := msg.CreatedAt.Format("02/01/2006 15:04:05")
	if msg.CreatedAt.IsZero() {
		createdTime = time.Now().Format("02/01/2006 15:04:05")
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="vi">
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; margin: 0; padding: 24px; color: #1e293b; }
    .card { max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05); }
    .header { background: #0f172a; padding: 24px; color: #ffffff; }
    .header h2 { margin: 0; font-size: 20px; font-weight: 600; }
    .header p { margin: 6px 0 0; font-size: 14px; color: #94a3b8; }
    .content { padding: 24px; }
    .info-table { width: 100%%; border-collapse: collapse; margin-bottom: 20px; }
    .info-table td { padding: 8px 0; font-size: 14px; vertical-align: top; }
    .info-label { width: 130px; font-weight: 600; color: #64748b; }
    .message-box { background: #f1f5f9; border-left: 4px solid #3b82f6; padding: 16px; border-radius: 6px; font-size: 15px; line-height: 1.6; white-space: pre-wrap; color: #0f172a; }
    .footer { padding: 16px 24px; background: #f8fafc; border-top: 1px solid #e2e8f0; font-size: 13px; color: #64748b; text-align: center; }
    .btn { display: inline-block; background: #2563eb; color: #ffffff !important; padding: 10px 20px; text-decoration: none; border-radius: 6px; font-weight: 500; font-size: 14px; margin-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
    <div class="header">
      <h2>📬 Tin nhắn liên hệ mới</h2>
      <p>Nhận từ website portfolio cá nhân</p>
    </div>
    <div class="content">
      <table class="info-table">
        <tr>
          <td class="info-label">Người gửi:</td>
          <td><strong>%s</strong></td>
        </tr>
        <tr>
          <td class="info-label">Email:</td>
          <td><a href="mailto:%s" style="color: #2563eb;">%s</a></td>
        </tr>
        <tr>
          <td class="info-label">Tiêu đề:</td>
          <td>%s</td>
        </tr>
        <tr>
          <td class="info-label">Thời gian:</td>
          <td>%s</td>
        </tr>
      </table>

      <div style="font-weight: 600; font-size: 14px; margin-bottom: 8px; color: #475569;">Nội dung tin nhắn:</div>
      <div class="message-box">%s</div>

      <div style="text-align: center;">
        <a class="btn" href="mailto:%s?subject=Re: %s">Phản hồi qua email</a>
      </div>
    </div>
    <div class="footer">
      Tin nhắn được gửi từ Portfolio Nguyễn Quốc Bảo • Hệ thống tự động
    </div>
  </div>
</body>
</html>`,
		html.EscapeString(fullName),
		html.EscapeString(msg.Email),
		html.EscapeString(msg.Email),
		html.EscapeString(msg.Subject),
		html.EscapeString(createdTime),
		html.EscapeString(msg.Message),
		html.EscapeString(msg.Email),
		html.EscapeString(msg.Subject),
	)
}

func (s *smtpEmailService) buildAutoReplyHTML(msg *models.ContactMessage, fullName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="vi">
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; margin: 0; padding: 24px; color: #1e293b; }
    .card { max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05); }
    .header { background: #0f172a; padding: 24px; color: #ffffff; text-align: center; }
    .header h2 { margin: 0; font-size: 20px; font-weight: 600; }
    .content { padding: 24px; font-size: 15px; line-height: 1.6; color: #334155; }
    .quote-box { background: #f8fafc; border: 1px solid #e2e8f0; padding: 14px; border-radius: 6px; font-size: 14px; color: #64748b; margin-top: 16px; }
    .footer { padding: 16px 24px; background: #f8fafc; border-top: 1px solid #e2e8f0; font-size: 13px; color: #94a3b8; text-align: center; }
  </style>
</head>
<body>
  <div class="card">
    <div class="header">
      <h2>✨ Cảm ơn bạn đã liên hệ</h2>
    </div>
    <div class="content">
      <p>Xin chào <strong>%s</strong>,</p>
      <p>Cảm ơn bạn đã gửi tin nhắn qua website portfolio của tôi. Tôi đã nhận được thông tin liên hệ của bạn và sẽ xem xét, phản hồi lại qua địa chỉ email <strong>%s</strong> trong thời gian sớm nhất.</p>

      <div class="quote-box">
        <div style="font-weight: 600; margin-bottom: 4px; color: #475569;">Trích dẫn nội dung đã gửi:</div>
        <div><em>"%s"</em></div>
      </div>

      <p style="margin-top: 24px;">Trân trọng,<br><strong>Nguyễn Quốc Bảo</strong></p>
    </div>
    <div class="footer">
      Đây là email xác nhận tự động. Vui lòng không trả lời trực tiếp email này.
    </div>
  </div>
</body>
</html>`,
		html.EscapeString(fullName),
		html.EscapeString(msg.Email),
		html.EscapeString(msg.Message),
	)
}
