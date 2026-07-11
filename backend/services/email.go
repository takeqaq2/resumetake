package services

import (
	cryptorand "crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// R56-B5: refuse to send SMTP credentials over unencrypted connections.
	// Mirrors the protection in smtp.PlainAuth.Start. The current code path
	// always uses tls.DialWithDialer so server.TLS is always true, but this
	// prevents a future configuration change (e.g. adding SMTP_SKIP_TLS for
	// local dev) from silently leaking credentials.
	if !server.TLS {
		return "", nil, errors.New("unencrypted connection: SMTP credentials require TLS")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	challenge := strings.TrimRight(strings.ToLower(string(fromServer)), ": ")
	switch challenge {
	case "username":
		return []byte(a.username), nil
	case "password":
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
}

func loginAuthFunc(username, password string) smtp.Auth {
	return &loginAuth{username: username, password: password}
}

func SendVerificationEmail(toEmail, code string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured — cannot send verification email")
	}
	if smtpFrom == "" {
		smtpFrom = smtpUser
	}

	// R49-B1: strip CR/LF from user-controlled header values to prevent SMTP
	// header injection. An attacker who controls toEmail (e.g. via the
	// send-code endpoint) could inject "\r\nBcc: victim@example.com" to
	// add hidden recipients, or "\r\n\r\n" to break out of headers and
	// inject a second message body. smtpFrom is env-controlled but sanitized
	// defensively for consistency. Only \r and \n need stripping — other
	// control characters are handled by the SMTP DATA writer.
	sanitizeHeader := func(s string) string {
		return strings.NewReplacer("\r", "", "\n", "").Replace(s)
	}
	toEmail = sanitizeHeader(toEmail)
	smtpFrom = sanitizeHeader(smtpFrom)

	// R28-L1: include required RFC 5322 headers (Date, From, To, Message-ID)
	// to avoid spam filters and comply with the standard. Message-ID uses the
	// app domain; Date uses RFC 1123Z format as required by RFC 5322 §3.6.1.
	dateHeader := "Date: " + time.Now().Format(time.RFC1123Z) + "\r\n"
	fromHeader := "From: ResumeTake <" + smtpFrom + ">\r\n"
	toHeader := "To: " + toEmail + "\r\n"
	subject := "Subject: ResumeTake Verification Code\r\n"
	// R31-1: add a CSPRNG suffix so concurrent sends in the same nanosecond
	// don't collide on Message-ID. Duplicate Message-IDs cause SMTP servers
	// and downstream filters to dedup legitimate verification emails.
	var rnd [4]byte
	if _, err := cryptorand.Read(rnd[:]); err != nil {
		return fmt.Errorf("failed to generate Message-ID randomness: %w", err)
	}
	// R52b-B4: derive Message-ID domain from SMTP_FROM so it aligns with
	// the From header (DMARC alignment). Previously hardcoded resume.takee.top,
	// which triggers spam filters if the deployment uses a different domain.
	msgDomain := "resume.takee.top"
	if at := strings.LastIndex(smtpFrom, "@"); at != -1 {
		msgDomain = smtpFrom[at+1:]
	}
	messageID := fmt.Sprintf("Message-ID: <%d-%x.resumetake@%s>\r\n", time.Now().UnixNano(), rnd[:], msgDomain)
	contentType := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n"
	body := fmt.Sprintf(`<html><body><div style="font-family:Arial,sans-serif;max-width:400px;margin:0 auto;padding:20px">
<h2 style="color:#4F46E5">ResumeTake</h2>
<p>Your verification code is:</p>
<div style="font-size:32px;font-weight:bold;color:#4F46E5;letter-spacing:8px;text-align:center;padding:20px;background:#F3F4F6;border-radius:8px">%s</div>
<p style="color:#6B7280;font-size:14px">This code expires in 5 minutes.</p>
</div></body></html>`, code)

	msg := []byte(dateHeader + fromHeader + toHeader + messageID + subject + contentType + body)

	addr := smtpHost + ":" + smtpPort
	auth := loginAuthFunc(smtpUser, smtpPass)

	tlsconfig := &tls.Config{ServerName: smtpHost}
	// Use a dialer with a timeout so an unresponsive SMTP server doesn't
	// hang the handler goroutine indefinitely (tls.Dial has no timeout).
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS dial: %w", err)
	}
	defer conn.Close()
	// Set an overall deadline for all SMTP operations (auth, mail, rcpt,
	// data, quit). Without this, a stalled SMTP server can hang each step
	// indefinitely, exhausting goroutines under heavy send-code traffic.
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = client.Mail(smtpFrom); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	// R39b-M6: w.Close() success means the message was accepted by the SMTP
	// server. client.Quit() sends the QUIT command to gracefully close the
	// connection, but if it fails (network blip, server already gone), the
	// email was still delivered — returning an error here would make the
	// caller show "email send failed" to the user when it actually succeeded.
	// Log the QUIT error for observability but don't propagate it.
	if err := client.Quit(); err != nil {
		// R54b-B2: omit the recipient email from the log — the rest of the
		// codebase hashes emails before logging (GDPR minimization). The
		// recipient is already recorded in the delivery log above; repeating
		// it here in plaintext leaks PII in the error path.
		log.Printf("[WARN] SMTP Quit failed (email already delivered): %v", err)
	}
	return nil
}

func GenerateVerificationCode() (string, error) {
	// Fail-closed: if the system CSPRNG is unavailable, return an error
	// instead of silently producing a predictable "000000" code that an
	// attacker could use to verify any email address.
	// Use rejection sampling to eliminate modulo bias: 2^32 = 4294967296,
	// and 4294 * 1000000 = 4294000000, so any value >= 4294000000 would
	// over-represent codes 0..967295. Reject those and redraw.
	for i := 0; i < 16; i++ {
		var b [4]byte
		if _, err := cryptorand.Read(b[:]); err != nil {
			return "", fmt.Errorf("failed to generate secure code: %w", err)
		}
		n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
		if n < 4294000000 {
			return fmt.Sprintf("%06d", n%1000000), nil
		}
	}
	// Extremely unlikely (probability < 1e-48 for 16 draws), but fail-closed
	// rather than falling back to a biased value.
	return "", fmt.Errorf("failed to generate unbiased secure code after 16 attempts")
}
