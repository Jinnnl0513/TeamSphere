package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/repository"
)

var ErrEmailDisabled = errors.New("email is not enabled")

// EmailService provides runtime email functionality using SMTP settings
// stored in the system_settings table.
// Fix #1: SMTP password is stored encrypted with AES-256-GCM.
type EmailService struct {
	settingsRepo  repository.SettingsRepository
	encryptionKey string // hex-encoded 32-byte key from config.yaml
}

func NewEmailService(settingsRepo repository.SettingsRepository, encryptionKey string) *EmailService {
	return &EmailService{settingsRepo: settingsRepo, encryptionKey: encryptionKey}
}

// EmailSettings holds the SMTP configuration read from system_settings.
// Password is always returned decrypted (for internal use) or masked (for API responses).
type EmailSettings struct {
	Enabled     bool
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string // decrypted plaintext password (never sent to API clients)
	FromAddress string
	FromName    string
}

// GetSettings reads email configuration from the system_settings table.
// The stored SMTP password is automatically decrypted if an encryption key is available.
func (s *EmailService) GetSettings(ctx context.Context) (*EmailSettings, error) {
	legacyEnabled, _ := s.settingsRepo.Get(ctx, "email.enabled")
	requiredEnabled, _ := s.settingsRepo.Get(ctx, "registration.email_required")
	host, _ := s.settingsRepo.Get(ctx, "email.smtp_host")
	portStr, _ := s.settingsRepo.Get(ctx, "email.smtp_port")
	username, _ := s.settingsRepo.Get(ctx, "email.username")
	encryptedPassword, _ := s.settingsRepo.Get(ctx, "email.password")
	fromAddr, _ := s.settingsRepo.Get(ctx, "email.from_address")
	fromName, _ := s.settingsRepo.Get(ctx, "email.from_name")

	enabled := requiredEnabled == "true" || legacyEnabled == "true"

	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Decrypt SMTP password if key is available
	password := encryptedPassword
	if s.encryptionKey != "" && encryptedPassword != "" {
		if decrypted, err := config.Decrypt(s.encryptionKey, encryptedPassword); err == nil {
			password = decrypted
		} else {
			// Graceful degradation: if decrypt fails, the stored value may be legacy plaintext
			slog.Warn("failed to decrypt email password, using stored value as-is", "error", err)
		}
	}

	return &EmailSettings{
		Enabled:     enabled,
		SMTPHost:    host,
		SMTPPort:    port,
		Username:    username,
		Password:    password,
		FromAddress: fromAddr,
		FromName:    fromName,
	}, nil
}

// UpdateSettings writes email configuration to the system_settings table.
// Fix #1: The SMTP password is encrypted with AES-256-GCM before storage.
func (s *EmailService) UpdateSettings(ctx context.Context, settings *EmailSettings) error {
	// Encrypt password before storing
	passwordToStore := settings.Password
	if s.encryptionKey != "" && settings.Password != "" {
		encrypted, err := config.Encrypt(s.encryptionKey, settings.Password)
		if err != nil {
			return fmt.Errorf("encrypt email password: %w", err)
		}
		passwordToStore = encrypted
	}

	pairs := map[string]string{
		"email.enabled":      fmt.Sprintf("%t", settings.Enabled),
		"email.smtp_host":    settings.SMTPHost,
		"email.smtp_port":    strconv.Itoa(settings.SMTPPort),
		"email.username":     settings.Username,
		"email.password":     passwordToStore,
		"email.from_address": settings.FromAddress,
		"email.from_name":    settings.FromName,
	}
	for k, v := range pairs {
		if err := s.settingsRepo.Set(ctx, k, v); err != nil {
			return fmt.Errorf("update setting %s: %w", k, err)
		}
	}
	return nil
}

// SendEmail sends a simple email using the current SMTP settings from system_settings.
// Returns an error if email is not enabled or sending fails.
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("get email settings: %w", err)
	}
	if !settings.Enabled {
		return ErrEmailDisabled
	}

	msgBytes, err := buildPlainTextMessage(settings.FromName, settings.FromAddress, to, subject, body)
	if err != nil {
		return fmt.Errorf("build email message: %w", err)
	}
	if err := sendSMTPMessage(settings.SMTPHost, settings.SMTPPort, settings.Username, settings.Password, settings.FromAddress, []string{to}, msgBytes); err != nil {
		slog.Error("failed to send email", "to", to, "error", err)
		return fmt.Errorf("send email: %w", err)
	}

	slog.Info("email sent", "to", to, "subject", subject)
	return nil
}

func sendSMTPMessage(host string, port int, username, password, from string, to []string, msg []byte) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	if port == 465 {
		return sendMailImplicitTLS(addr, host, username, password, from, to, msg)
	}

	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return smtp.SendMail(addr, auth, from, to, msg)
}
func buildPlainTextMessage(fromName, fromAddress, to, subject, body string) ([]byte, error) {
	encodedSubject := mime.BEncoding.Encode("UTF-8", subject)
	fromHeader := fromAddress
	if fromName != "" {
		encodedName := mime.BEncoding.Encode("UTF-8", fromName)
		fromHeader = fmt.Sprintf("%s <%s>", encodedName, fromAddress)
	}

	var bodyBuf bytes.Buffer
	qp := quotedprintable.NewWriter(&bodyBuf)
	if _, err := qp.Write([]byte(body)); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}

	messageIDDomain := "localhost"
	if i := strings.IndexByte(fromAddress, '@'); i >= 0 && i+1 < len(fromAddress) {
		messageIDDomain = fromAddress[i+1:]
	}

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString(fmt.Sprintf("Message-ID: <%d@%s>\r\n", time.Now().UnixNano(), messageIDDomain))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString("\r\n")
	msg.Write(bodyBuf.Bytes())
	return msg.Bytes(), nil
}

func sendMailImplicitTLS(addr, host, username, password, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
