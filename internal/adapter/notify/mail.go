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
	b.WriteString("Message-ID: <" + id.NewString() + "@sunshinefacility.com>\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(strings.ReplaceAll(m.Body, "\n", "\r\n"))

	addr := net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
	var auth smtp.Auth
	if s.User != "" {
		auth = smtp.PlainAuth("", s.User, s.Password, s.Host)
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

// Log only records that a mail would have been sent.
type Log struct{ L *slog.Logger }

func (l Log) Send(_ context.Context, m app.Mail) error {
	l.L.Info("mail not sent (notifications disabled)", "to_count", len(m.To), "subject", m.Subject)
	return nil
}
