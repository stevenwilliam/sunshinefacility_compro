package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/certificate"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/params"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
)

type Jobs struct {
	store    Store
	storage  Storage
	mailer   Mailer
	settings *Settings
	adminURL string
	log      *slog.Logger
	now      func() time.Time
}

func NewJobs(store Store, storage Storage, mailer Mailer, settings *Settings, adminURL string, log *slog.Logger, now func() time.Time) *Jobs {
	return &Jobs{store: store, storage: storage, mailer: mailer, settings: settings, adminURL: adminURL, log: log, now: now}
}

// backoff after the n-th failed attempt. After ten failures a mail stops
// being retried; it stays in the outbox with its last error for a person to
// see (CMS → Outgoing mail).
func backoff(attempt int) time.Duration {
	steps := []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour}
	if attempt-1 < len(steps) {
		return steps[attempt-1]
	}
	return 24 * time.Hour
}

const maxMailAttempts = 10

// DeliverMail sends what is due. Rows are claimed FOR UPDATE SKIP LOCKED, so
// two workers never send the same mail twice.
func (j *Jobs) DeliverMail(ctx context.Context) (sent, failed int, err error) {
	from := j.settings.String(ctx, "notify.from", "noreply@sunshinefacility.com")
	err = j.store.InTx(ctx, func(s Store) error {
		due, err := s.DueMail(ctx, 20)
		if err != nil {
			return err
		}
		for _, m := range due {
			sendErr := j.mailer.Send(ctx, Mail{From: from, To: m.Recipients, Subject: m.Subject, Body: m.Body})
			if sendErr == nil {
				if err := s.MarkMailSent(ctx, m.ID); err != nil {
					return err
				}
				sent++
				continue
			}
			failed++
			next := j.now().Add(backoff(m.Attempts + 1))
			if m.Attempts+1 >= maxMailAttempts {
				next = j.now().AddDate(100, 0, 0)
			}
			j.log.Warn("mail delivery failed", "outbox_id", m.ID, "attempt", m.Attempts+1, "error", sendErr.Error())
			if err := s.MarkMailFailed(ctx, m.ID, truncate(sendErr.Error(), 500), next); err != nil {
				return err
			}
		}
		return nil
	})
	return sent, failed, err
}

// CertificateReminders — BR-3.4. One mail per certificate per expiry date:
// the ledger insert and the mail are in one transaction.
func (j *Jobs) CertificateReminders(ctx context.Context) (int, error) {
	today := calendar.Today(j.now())
	days := j.settings.Int(ctx, "certificates.reminder_days", 60)
	to := params.RecipientList(j.settings.String(ctx, "notify.certificate_to", ""))
	if len(to) == 0 {
		return 0, nil
	}
	certs, err := j.store.CertificatesForReminder(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, c := range certs {
		if c.ValidUntil == "" {
			continue
		}
		vu, err := calendar.ParseDate(c.ValidUntil)
		if err != nil || !certificate.ReminderDue(&vu, today, days) {
			continue
		}
		left := today.DaysUntil(vu)
		err = j.store.InTx(ctx, func(s Store) error {
			fresh, err := s.RecordCertificateReminder(ctx, c.ID, c.ValidUntil)
			if err != nil || !fresh {
				return err
			}
			body := fmt.Sprintf(`The certificate "%s" expires on %s, in %d day(s).

When it expires it disappears from the website automatically, and the CMS will not let anyone publish it again until the renewed certificate is uploaded and its date updated.

Update it in the CMS: %s/r/certificates
`, c.Name, vu.Time().Format("2 January 2006"), left, j.adminURL)
			n++
			return s.EnqueueMail(ctx, OutboxMail{ID: id.NewString(), Kind: "certificate_expiry", RefID: c.ID, Recipients: to,
				Subject: "Certificate expiring: " + c.Name + " on " + vu.Time().Format("2 Jan 2006"), Body: body, NextAt: j.now()})
		})
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// PurgeApplications — BR-6.3. Applications and their CVs older than the
// retention parameter are deleted, each one audited as a system action.
func (j *Jobs) PurgeApplications(ctx context.Context) (int, error) {
	months := j.settings.Int(ctx, "careers.retention_months", 12)
	cutoff := j.now().AddDate(0, -months, 0)
	old, err := j.store.ApplicationsOlderThan(ctx, cutoff)
	if err != nil {
		return 0, err
	}
	n := 0
	actor := System("purge-applications")
	for _, a := range old {
		err := j.store.InTx(ctx, func(s Store) error {
			if err := s.DeleteApplication(ctx, a.ID); err != nil {
				return err
			}
			return audit(ctx, s, actor, "delete", "application", a.ID,
				fmt.Sprintf("Deleted after %d months (retention)", months), map[string]any{"receivedAt": a.CreatedAt, "position": a.Position}, nil)
		})
		if err != nil {
			return n, err
		}
		// The row goes first, so nothing can serve the CV any more. If the
		// object delete fails, the key is logged for manual removal; there is
		// no automatic retry for it.
		if err := j.storage.Delete(ctx, a.CVKey); err != nil {
			j.log.Error("cv delete failed after purge", "application_id", a.ID, "error", err.Error())
		}
		n++
	}
	return n, nil
}

// RunWorker runs the mail worker until ctx ends. Reminders and the purge
// are daily systemd timers instead (deploy/), so a restart cannot skip them.
func (j *Jobs) RunWorker(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		if _, _, err := j.DeliverMail(ctx); err != nil && ctx.Err() == nil {
			j.log.Error("mail worker", "error", err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
