package mailer

import (
	"bytes"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"1001-twacc-chat/internal/config"
)

// SMTPMailer sends transactional emails through an SMTP server.
type SMTPMailer struct {
	host        string
	port        int
	username    string
	password    string
	fromAddress string
	fromName    string
}

// NewSMTPMailer builds an SMTP-backed mailer from config.
func NewSMTPMailer(cfg config.MailConfig) *SMTPMailer {
	return &SMTPMailer{
		host:        cfg.SMTPHost,
		port:        cfg.SMTPPort,
		username:    cfg.Username,
		password:    cfg.Password,
		fromAddress: cfg.SenderAddress(),
		fromName:    cfg.FromName,
	}
}

// SendUserInvitation sends a registration invitation email.
func (m *SMTPMailer) SendUserInvitation(email, inviteURL string, expiresAt time.Time) error {
	if m == nil {
		return fmt.Errorf("smtp mailer unavailable")
	}
	to := strings.TrimSpace(email)
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid invitation email address: %w", err)
	}
	from := strings.TrimSpace(m.fromAddress)
	if from == "" {
		from = strings.TrimSpace(m.username)
	}
	fromAddress := mail.Address{Name: strings.TrimSpace(m.fromName), Address: from}
	subject := "TWACC 聊天系統註冊邀請"
	body := fmt.Sprintf("您好，\n\n請點選以下連結完成 TWACC 聊天系統帳號設定：\n%s\n\n此邀請將於 %s 到期。\n\n若您沒有申請此邀請，請忽略此信。\n",
		inviteURL,
		expiresAt.In(time.Local).Format("2006-01-02 15:04:05"),
	)

	var msg bytes.Buffer
	msg.WriteString("From: " + fromAddress.String() + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := net.JoinHostPort(m.host, fmt.Sprintf("%d", m.port))
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg.Bytes()); err != nil {
		return fmt.Errorf("send invitation email: %w", err)
	}
	return nil
}
