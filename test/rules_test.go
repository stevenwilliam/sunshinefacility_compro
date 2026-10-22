package test

import (
	"strings"
	"testing"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
)

// BR-3.2 — refused by the API with the reason, and by the database when the
// API is bypassed.
func TestBR3_2_ExpiredCertificateCannotBePublished(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	halal := e.id(`SELECT id::text FROM certificates WHERE name = 'Halal'`)
	var out struct {
		Error struct {
			Code   string
			Fields map[string]string
		}
	}
	ed.do("PUT", "/api/v1/admin/r/certificates/"+halal, map[string]any{"published": true}, &out, 422)
	if !strings.Contains(out.Error.Fields["published"], "expired on 25 Aug 2025") {
		t.Fatalf("refusal does not name the date: %+v", out)
	}
	if err := e.gdb.Exec(`UPDATE certificates SET published = true WHERE id = ?::uuid`, halal).Error; err == nil {
		t.Fatal("the database accepted publishing an expired certificate")
	}
	// Renewing it (new date) makes it publishable again.
	ed.do("PUT", "/api/v1/admin/r/certificates/"+halal, map[string]any{"valid_until": "2028-08-25", "published": true}, nil, 200)
}

// BR-3.3 — a published certificate disappears from the site the day after
// it expires, with no write.
func TestBR3_3_ExpiredCertificateHiddenWithoutAWrite(t *testing.T) {
	e := fresh(t)
	tomorrow := calendar.Today(e.now).AddDays(1).String()
	if err := e.gdb.Exec(`UPDATE certificates SET valid_until = ?::date WHERE name = 'SLHS'`, tomorrow).Error; err != nil {
		t.Fatal(err)
	}
	names := func() string {
		var b struct{ Certificates []struct{ Name string } }
		e.anon().do("GET", "/api/v1/public/bundle", nil, &b, 200)
		var n []string
		for _, c := range b.Certificates {
			n = append(n, c.Name)
		}
		return strings.Join(n, ",")
	}
	if !strings.Contains(names(), "SLHS") {
		t.Fatal("SLHS hidden before it expired")
	}
	e.now = e.now.Add(48 * time.Hour) // the day after tomorrow, in Jakarta
	if strings.Contains(names(), "SLHS") {
		t.Fatal("an expired certificate is still on the public site")
	}
	if strings.Contains(names(), "Halal") {
		t.Fatal("the unpublished halal certificate is in the bundle")
	}
}

// BR-3.4 — one reminder per certificate per expiry date.
func TestBR3_4_ReminderOncePerExpiry(t *testing.T) {
	e := fresh(t)
	in30 := calendar.Today(e.now).AddDays(30).String()
	if err := e.gdb.Exec(`UPDATE certificates SET valid_until = ?::date WHERE name = 'ISO 22000:2018'`, in30).Error; err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	n, err := e.jobs.CertificateReminders(ctx)
	if err != nil || n != 1 {
		t.Fatalf("first run queued %d (%v), want 1", n, err)
	}
	if n, _ := e.jobs.CertificateReminders(ctx); n != 0 {
		t.Fatalf("second run queued %d, want 0", n)
	}
	if _, _, err := e.jobs.DeliverMail(ctx); err != nil {
		t.Fatal(err)
	}
	if len(e.mailer.sent) != 1 || !strings.Contains(e.mailer.sent[0].Subject, "ISO 22000:2018") {
		t.Fatalf("mail: %+v", e.mailer.sent)
	}
}

// BR-4.3 — an image in use cannot be deleted; the message says how many use it.
func TestBR4_3_MediaInUseCannotBeDeleted(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	logo := e.id(`SELECT id::text FROM media WHERE slug = 'logo/danone'`)
	var out struct {
		Error struct{ Code, Message string }
	}
	ed.do("DELETE", "/api/v1/admin/media/"+logo, nil, &out, 409)
	if out.Error.Code != "IN_USE" || !strings.Contains(out.Error.Message, "used in 1 item") {
		t.Fatalf("got %+v", out)
	}
	if err := e.gdb.Exec(`DELETE FROM media WHERE id = ?::uuid`, logo).Error; err == nil {
		t.Fatal("the database let an in-use image be deleted")
	}
	// An unused image can go.
	unused := e.id(`SELECT id::text FROM media WHERE slug = 'photo/gardening'`)
	ed.do("DELETE", "/api/v1/admin/media/"+unused, nil, nil, 200)
}

// BR-2.2 / BR-2.3 — content edits clear the flag; publishing does not.
func TestBR2_PlaceholderFlag(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	tid := e.id(`SELECT id::text FROM testimonials ORDER BY sort LIMIT 1`)
	var r map[string]any
	ed.do("PUT", "/api/v1/admin/r/testimonials/"+tid, map[string]any{"published": true, "sort": 5, "is_placeholder": true}, &r, 200)
	if r["is_placeholder"] != true {
		t.Fatal("publishing cleared the placeholder flag")
	}
	ed.do("PUT", "/api/v1/admin/r/testimonials/"+tid, map[string]any{"quote_en": "They answered in ten minutes.", "is_placeholder": true}, &r, 200)
	if r["is_placeholder"] != false {
		t.Fatal("a content edit did not clear the placeholder flag")
	}
	ed.do("PUT", "/api/v1/admin/r/testimonials/"+tid, map[string]any{"is_placeholder": true}, &r, 200)
	if r["is_placeholder"] != true {
		t.Fatal("an explicit re-flag was ignored")
	}
}

// BR-2.4 — unpublished placeholder quotes show; real unpublished ones never leak.
func TestBR2_4_TestimonialPlaceholdersInBundle(t *testing.T) {
	e := fresh(t)
	var b struct {
		Testimonials            []any
		TestimonialPlaceholders []struct{ Name string }
	}
	e.anon().do("GET", "/api/v1/public/bundle", nil, &b, 200)
	if len(b.Testimonials) != 0 || len(b.TestimonialPlaceholders) != 3 {
		t.Fatalf("published %d, placeholders %d", len(b.Testimonials), len(b.TestimonialPlaceholders))
	}
	if err := e.gdb.Exec(`UPDATE testimonials SET is_placeholder = false, quote_en = 'Real, not yet approved' WHERE sort = 1`).Error; err != nil {
		t.Fatal(err)
	}
	e.public.Invalidate()
	var s string
	e.anon().do("GET", "/api/v1/public/bundle", nil, &s, 200)
	if strings.Contains(s, "Real, not yet approved") {
		t.Fatal("an unpublished real testimonial leaked into the public bundle")
	}
}

// BR-11.1 — a colour pair that fails AA is refused, with the ratio.
func TestBR11_1_ContrastRefused(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	id := e.id(`SELECT id::text FROM service_lines WHERE key = 'remote'`)
	var out struct {
		Error struct{ Fields map[string]string }
	}
	ed.do("PUT", "/api/v1/admin/r/services/"+id, map[string]any{"color_on_ink": "#FFFFFF"}, &out, 422)
	if !strings.Contains(out.Error.Fields["color_on_ink"], ":1") {
		t.Fatalf("no ratio in the refusal: %+v", out)
	}
}

// BR-1.3 — the public bundle carries no system parameters, no recipients.
func TestBR1_3_BundleLeaksNothingPrivate(t *testing.T) {
	e := fresh(t)
	var s string
	e.anon().do("GET", "/api/v1/public/bundle", nil, &s, 200)
	for _, leak := range []string{"itdept.sfg@gmail.com", "noreply@", "retention_months", "password_min_length", "Halal certification"} {
		if strings.Contains(s, leak) {
			t.Errorf("bundle contains %q", leak)
		}
	}
}

// Parameters: values are typed; system groups are superadmin-only.
func TestParametersTypedAndScoped(t *testing.T) {
	e := fresh(t)
	su := e.user("superadmin")
	ed := e.user("editor")
	ret := e.id(`SELECT id::text FROM parameters WHERE grp = 'careers' AND key = 'retention_months'`)
	var out struct {
		Error struct{ Fields map[string]string }
	}
	su.do("PUT", "/api/v1/admin/r/parameters/"+ret, map[string]any{"value_en": "twelve"}, &out, 422)
	su.do("PUT", "/api/v1/admin/r/parameters/"+ret, map[string]any{"value_en": "18"}, nil, 200)
	ed.do("PUT", "/api/v1/admin/r/parameters/"+ret, map[string]any{"value_en": "1"}, nil, 404) // not visible to an editor
	var list struct{ Rows []map[string]any }
	ed.do("GET", "/api/v1/admin/r/parameters", nil, &list, 200)
	for _, r := range list.Rows {
		if r["grp"] == "notify" || r["grp"] == "auth" {
			t.Fatalf("editor sees system parameter %v", r["key"])
		}
	}
	// A closed group's key cannot be deleted; an option can.
	phone := e.id(`SELECT id::text FROM parameters WHERE grp = 'contact' AND key = 'phone'`)
	ed.do("DELETE", "/api/v1/admin/r/parameters/"+phone, nil, nil, 409)
	inter := e.id(`SELECT id::text FROM parameters WHERE grp = 'job_types' AND key = 'in'`)
	ed.do("DELETE", "/api/v1/admin/r/parameters/"+inter, nil, nil, 200)
	// A location used by an opening cannot be deleted (FK).
	tg := e.id(`SELECT id::text FROM parameters WHERE grp = 'locations' AND key = 'tangerang'`)
	ed.do("DELETE", "/api/v1/admin/r/parameters/"+tg, nil, nil, 409)
}

// BR-10.1 — every change writes an audit row; the trail cannot be edited.
func TestBR10_1_AuditTrail(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	id := e.id(`SELECT id::text FROM clients WHERE name = 'Danone'`)
	ed.do("PUT", "/api/v1/admin/r/clients/"+id, map[string]any{"name": "Danone Indonesia"}, nil, 200)
	if n := e.count(`SELECT count(*) FROM audit_log WHERE entity = 'clients' AND entity_id = ? AND actor_email = 'editor@sfg.test'
		AND before->>'name' = 'Danone' AND after->>'name' = 'Danone Indonesia'`, id); n != 1 {
		t.Fatalf("audit rows = %d, want 1", n)
	}
	if err := e.gdb.Exec(`DELETE FROM audit_log`).Error; err == nil {
		t.Fatal("the audit log accepted a DELETE")
	}
	if err := e.gdb.Exec(`UPDATE audit_log SET summary = 'x'`).Error; err == nil {
		t.Fatal("the audit log accepted an UPDATE")
	}
}

// BR-6.3 — applications older than the retention period are deleted with
// their CVs; newer ones stay.
func TestBR6_3_RetentionPurge(t *testing.T) {
	e := fresh(t)
	ins := func(age time.Duration, key string) {
		if err := e.gdb.Exec(`INSERT INTO career_applications (id, created_at, name, phone, email, position, cv_key, cv_filename, cv_mime, cv_bytes, lang)
			VALUES (gen_random_uuid(), ?, 'A', '0812345678', 'a@x.com', 'Cook', ?, 'cv.pdf', 'application/pdf', 10, 'en')`, e.now.Add(-age), key).Error; err != nil {
			t.Fatal(err)
		}
		_ = e.storage.Put(t.Context(), key, strings.NewReader("%PDF-"), 5, "application/pdf")
	}
	ins(400*24*time.Hour, "cv/old.pdf")
	ins(30*24*time.Hour, "cv/new.pdf")
	n, err := e.jobs.PurgeApplications(t.Context())
	if err != nil || n != 1 {
		t.Fatalf("purged %d (%v), want 1", n, err)
	}
	if e.storage.has("cv/old.pdf") || !e.storage.has("cv/new.pdf") {
		t.Fatal("wrong CV removed")
	}
	if c := e.count(`SELECT count(*) FROM audit_log WHERE action = 'delete' AND entity = 'application' AND actor_email = 'system:purge-applications'`); c != 1 {
		t.Fatalf("purge audit rows = %d", c)
	}
}

// Outbox: a failing relay is retried with back-off; the submission survives.
func TestBR5_3_MailRetriedNotLost(t *testing.T) {
	e := fresh(t)
	e.anon().do("POST", "/api/v1/public/quotation", quote("x@contoh.co.id"), nil, 200)
	e.mailer.fail = true
	if _, failed, _ := e.jobs.DeliverMail(t.Context()); failed != 1 {
		t.Fatalf("failed = %d", failed)
	}
	if e.count(`SELECT count(*) FROM quotation_requests`) != 1 {
		t.Fatal("the submission was lost with the mail")
	}
	e.mailer.fail = false
	e.now = e.now.Add(2 * time.Minute)
	if sent, _, _ := e.jobs.DeliverMail(t.Context()); sent != 1 {
		t.Fatalf("retry sent %d", sent)
	}
}
