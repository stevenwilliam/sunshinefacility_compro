// Package notify implements app.Mailer. SMTP is the real one (mailpit in
// development, the production relay later); Log is the second
// implementation, used when SFG_NOTIFY_ENABLED=false (CLAUDE.md §9: every
// outbound integration has a port and at least two implementations).
package notify

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
)

type SMTP struct {
	Host     string
	Port     int
	User     string
	Password string
}

var _ app.Mailer = SMTP{}

// headerSafe refuses a CR or LF in anything that becomes a header: a
// visitor's company name must not be able to add a Bcc.
func headerSafe(s string) error {
	if strings.ContainsAny(s, "\r\n") {
		return fmt.Errorf("header value contains a line break")
	}
	return nil
}

func (s SMTP) Send(ctx context.Context, m app.Mail) error {
	for _, v := range append([]string{m.From, m.Subject}, m.To...) {
		if err := headerSafe(v); err != nil {
			return err
		}
	}
	var b strings.Builder
	b.WriteString("From: " + m.From + "\r\n")
	b.WriteString("To: " + strings.Join(m.To, ", ") + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", m.Subject) + "\r\n")
	b.WriteString("Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("Message-ID: <" + id.NewString() + "@" + messageIDDomain(m.From) + ">\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(strings.ReplaceAll(m.Body, "\n", "\r\n"))

	addr := net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
	var auth smtp.Auth
	if s.User != "" {
		auth = &autoAuth{user: s.User, password: s.Password, host: s.Host}
	}
	done := make(chan error, 1)
	go func() { done <- smtp.SendMail(addr, auth, m.From, m.To, []byte(b.String())) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("smtp: timed out after 30s")
	}
}

// autoAuth uses PLAIN when the server offers it and LOGIN otherwise:
// Office 365 advertises only "AUTH LOGIN XOAUTH2". Like smtp.PlainAuth it
// refuses to send credentials over an unencrypted link except to localhost.
type autoAuth struct {
	user, password, host string
	plain                smtp.Auth
}

func (a *autoAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	for _, m := range server.Auth {
		if m == "PLAIN" {
			a.plain = smtp.PlainAuth("", a.user, a.password, a.host)
			return a.plain.Start(server)
		}
	}
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, fmt.Errorf("smtp: refusing AUTH LOGIN over an unencrypted connection")
	}
	if server.Name != a.host {
		return "", nil, fmt.Errorf("smtp: wrong host name %q", server.Name)
	}
	return "LOGIN", nil, nil
}

func (a *autoAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if a.plain != nil {
		return a.plain.Next(fromServer, more)
	}
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(string(fromServer))) {
	case "username:":
		return []byte(a.user), nil
	case "password:":
		return []byte(a.password), nil
	}
	return nil, fmt.Errorf("smtp: unexpected AUTH LOGIN challenge %q", fromServer)
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}

// Log only records that a mail would have been sent.
type Log struct{ L *slog.Logger }

func (l Log) Send(_ context.Context, m app.Mail) error {
	l.L.Info("mail not sent (notifications disabled)", "to_count", len(m.To), "subject", m.Subject)
	return nil
}

// messageIDDomain takes the Message-ID right-hand side from the sender
// address, so it follows the notify.from parameter rather than a constant.
func messageIDDomain(from string) string {
	if i := strings.LastIndex(from, "@"); i >= 0 && i < len(from)-1 {
		return strings.Trim(from[i+1:], "> ")
	}
	return "localhost"
}
