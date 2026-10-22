package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/filetype"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/forms"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/params"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/sanitize"
)

// Submissions handles the two public forms and the read-only inbox.
type Submissions struct {
	store    Store
	storage  Storage
	settings *Settings
	public   *Public
	adminURL string
	now      func() time.Time
}

func NewSubmissions(store Store, storage Storage, settings *Settings, public *Public, adminURL string, now func() time.Time) *Submissions {
	return &Submissions{store: store, storage: storage, settings: settings, public: public, adminURL: adminURL, now: now}
}

// Envelope is what arrives with every public form besides its fields.
type Envelope struct {
	IP        string
	UserAgent string
	// Honeypot is a field hidden from people; a bot fills it.
	Honeypot string
	// ElapsedMS is how long the form was open before it was sent.
	ElapsedMS int64
}

// ErrDiscarded marks a submission accepted at the door and dropped as a
// bot. The handler answers exactly as it does for a real one, so the bot
// learns nothing (BR-5.4).
var ErrDiscarded = errors.New("discarded as automated")

// guard applies BR-5.4: honeypot, minimum fill time, rate limits.
func (s *Submissions) guard(ctx context.Context, table string, env Envelope, email string) error {
	if strings.TrimSpace(env.Honeypot) != "" {
		return ErrDiscarded
	}
	minMS := int64(s.settings.IntAllowZero(ctx, "forms.min_fill_seconds", 3)) * 1000
	if env.ElapsedMS < minMS {
		// A person can be this fast with autofill, so this is a visible
		// refusal with a reason, not a silent drop.
		return apierror.Validation("too_fast", map[string]string{"_form": "too_fast"})
	}
	since := s.now().Add(-time.Hour)
	if n, err := s.store.CountRecentSubmissions(ctx, table, "ip", env.IP, since); err != nil {
		return err
	} else if n >= s.settings.Int(ctx, "forms.per_ip_per_hour", 10) {
		return apierror.New(apierror.CodeRateLimited, "rate")
	}
	if email != "" {
		if n, err := s.store.CountRecentSubmissions(ctx, table, "email", strings.ToLower(email), since); err != nil {
			return err
		} else if n >= s.settings.Int(ctx, "forms.per_email_per_hour", 5) {
			return apierror.New(apierror.CodeRateLimited, "rate")
		}
	}
	return nil
}

func fieldErrors(errs forms.Errors) error {
	return apierror.Validation("check", map[string]string(errs))
}

// SubmitQuotation — BR-5.1…BR-5.4. Returns the visitor's first name for the
// thank-you line.
func (s *Submissions) SubmitQuotation(ctx context.Context, in forms.QuotationInput, env Envelope) (string, error) {
	email, _ := sanitize.Email(in.Email)
	if err := s.guard(ctx, "quotation_requests", env, email); err != nil {
		return "", err
	}
	b, _, err := s.public.Get(ctx)
	if err != nil {
		return "", err
	}
	var types, lines []string
	typeLabel := map[string][2]string{}
	for _, t := range b.BuildingTypes {
		types = append(types, t.Key)
		typeLabel[t.Key] = t.Label
	}
	lineTitle := map[string]string{}
	for _, l := range b.Lines {
		lines = append(lines, l.Key)
		lineTitle[l.Key] = l.Title[0]
	}
	q, errs := forms.ValidateQuotation(in, types, lines)
	if errs != nil {
		return "", fieldErrors(errs)
	}
	lbl := typeLabel[q.BuildingType][0]
	if q.Lang == "id" {
		lbl = typeLabel[q.BuildingType][1]
	}
	rec := QuotationRecord{ID: id.NewString(), CreatedAt: s.now().UTC(), Name: q.Name, Company: q.Company,
		Email: strings.ToLower(q.Email), Phone: q.Phone, BuildingTypeKey: q.BuildingType, BuildingType: lbl,
		AreaM2: q.AreaM2, Location: q.Location, Services: q.Services, Message: q.Message, Lang: q.Lang,
		IP: env.IP, UserAgent: truncate(env.UserAgent, 300)}

	var svc []string
	for _, k := range q.Services {
		svc = append(svc, lineTitle[k])
	}
	area := "—"
	if q.AreaM2 != nil {
		area = fmt.Sprintf("%d m²", *q.AreaM2)
	}
	body := fmt.Sprintf(`A new quotation request arrived on www.sunshinefacility.com.

Name:          %s
Company:       %s
Email:         %s
Phone:         %s
Building type: %s
Area:          %s
Location:      %s
Services:      %s
Language:      %s

Message:
%s

Open it in the CMS: %s/inbox/quotations/%s
`, rec.Name, rec.Company, rec.Email, rec.Phone, typeLabel[q.BuildingType][0], area, rec.Location,
		strings.Join(svc, ", "), strings.ToUpper(rec.Lang), orDash(rec.Message), s.adminURL, rec.ID)

	to := params.RecipientList(s.settings.String(ctx, "notify.quotation_to", ""))
	err = s.store.InTx(ctx, func(st Store) error {
		if err := st.InsertQuotation(ctx, rec); err != nil {
			return err
		}
		if len(to) == 0 {
			return nil // nobody to tell; the CMS inbox still has it
		}
		return st.EnqueueMail(ctx, OutboxMail{ID: id.NewString(), Kind: "quotation", RefID: rec.ID, Recipients: to,
			Subject: "Quotation request: " + oneLine(rec.Company) + " (" + oneLine(rec.Location) + ")", Body: body, NextAt: s.now()})
	})
	if err != nil {
		return "", err
	}
	return forms.FirstName(rec.Name), nil
}

// CV is the uploaded file.
type CV struct {
	Filename string
	Data     []byte
}

// SubmitApplication — BR-6.1, BR-6.2.
func (s *Submissions) SubmitApplication(ctx context.Context, in forms.ApplicationInput, cv *CV, env Envelope) error {
	email, _ := sanitize.Email(in.Email)
	if err := s.guard(ctx, "career_applications", env, email); err != nil {
		return err
	}
	b, _, err := s.public.Get(ctx)
	if err != nil {
		return err
	}
	var openings []string
	title := map[string][2]string{}
	for _, o := range b.Openings {
		openings = append(openings, o.ID)
		title[o.ID] = o.Title
	}
	a, errs := forms.ValidateApplication(in, openings)
	if errs == nil {
		errs = forms.Errors{}
	}
	maxMB := s.settings.Int(ctx, "careers.cv_max_mb", 5)
	var mime string
	switch {
	case cv == nil || len(cv.Data) == 0:
		errs["cv"] = forms.ErrRequired
	case int64(len(cv.Data)) > int64(maxMB)<<20:
		errs["cv"] = "file_size"
	default:
		var ok bool
		if mime, ok = filetype.Document(cv.Data); !ok {
			errs["cv"] = "file_type"
		}
	}
	if len(errs) > 0 {
		return fieldErrors(errs)
	}

	position := b.Text["careers.form.open_application"]
	if a.OpeningID != "" {
		position = title[a.OpeningID]
	}
	pos := position[0]
	if a.Lang == "id" {
		pos = position[1]
	}
	rec := ApplicationRecord{ID: id.NewString(), CreatedAt: s.now().UTC(), Name: a.Name, Phone: a.Phone,
		Email: strings.ToLower(a.Email), Position: pos, OpeningID: a.OpeningID, Message: a.Message,
		CVMime: mime, CVBytes: int64(len(cv.Data)), Lang: a.Lang, IP: env.IP, UserAgent: truncate(env.UserAgent, 300)}
	// The stored name is ours: the visitor's file name is kept only as a
	// label, sanitised, and the extension comes from the sniffed type.
	rec.CVFilename = sanitize.Filename(strings.TrimSuffix(cv.Filename, extOf(cv.Filename))) + filetype.Extension(mime)
	rec.CVKey = "cv/" + rec.ID + filetype.Extension(mime)
	if err := s.storage.Put(ctx, rec.CVKey, bytes.NewReader(cv.Data), int64(len(cv.Data)), mime); err != nil {
		return err
	}

	body := fmt.Sprintf(`A new job application arrived on www.sunshinefacility.com.

Name:      %s
Phone:     %s
Email:     %s
Position:  %s
CV:        %s (%d KB)
Language:  %s

Message:
%s

The CV is not attached: it is personal data and stays in the CMS.
Open it in the CMS: %s/inbox/applications/%s
`, rec.Name, rec.Phone, rec.Email, position[0], rec.CVFilename, (rec.CVBytes+1023)/1024, strings.ToUpper(rec.Lang),
		orDash(rec.Message), s.adminURL, rec.ID)

	to := params.RecipientList(s.settings.String(ctx, "notify.career_to", ""))
	err = s.store.InTx(ctx, func(st Store) error {
		if err := st.InsertApplication(ctx, rec); err != nil {
			return err
		}
		if len(to) == 0 {
			return nil
		}
		return st.EnqueueMail(ctx, OutboxMail{ID: id.NewString(), Kind: "application", RefID: rec.ID, Recipients: to,
			Subject: "Job application: " + oneLine(position[0]) + " — " + oneLine(rec.Name), Body: body, NextAt: s.now()})
	})
	if err != nil {
		_ = s.storage.Delete(ctx, rec.CVKey) // no orphaned personal data
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// Inbox (BR-7) — read-only.
// ---------------------------------------------------------------------------

func (s *Submissions) Quotations(ctx context.Context, actor Actor, q ListQuery) ([]QuotationRecord, int, error) {
	if !access.Can(access.Role(actor.Role), access.QuotationsRead) {
		return nil, 0, apierror.New(apierror.CodeForbidden, "You do not have access to quotation requests.")
	}
	return s.store.ListQuotations(ctx, q)
}

func (s *Submissions) Quotation(ctx context.Context, actor Actor, qid string) (*QuotationRecord, error) {
	if !access.Can(access.Role(actor.Role), access.QuotationsRead) {
		return nil, apierror.New(apierror.CodeForbidden, "You do not have access to quotation requests.")
	}
	r, err := s.store.GetQuotation(ctx, qid)
	if errors.Is(err, ErrNotFound) {
		return nil, apierror.NotFound("Quotation request")
	}
	return r, err
}

func (s *Submissions) Applications(ctx context.Context, actor Actor, q ListQuery) ([]ApplicationRecord, int, error) {
	if !access.Can(access.Role(actor.Role), access.ApplicationsRead) {
		return nil, 0, apierror.New(apierror.CodeForbidden, "You do not have access to job applications.")
	}
	return s.store.ListApplications(ctx, q)
}

func (s *Submissions) Application(ctx context.Context, actor Actor, aid string) (*ApplicationRecord, error) {
	if !access.Can(access.Role(actor.Role), access.ApplicationsRead) {
		return nil, apierror.New(apierror.CodeForbidden, "You do not have access to job applications.")
	}
	r, err := s.store.GetApplication(ctx, aid)
	if errors.Is(err, ErrNotFound) {
		return nil, apierror.NotFound("Application")
	}
	return r, err
}

// CVLink — BR-7.2: a five-minute link, and the download is audited.
func (s *Submissions) CVLink(ctx context.Context, actor Actor, aid string) (string, error) {
	a, err := s.Application(ctx, actor, aid)
	if err != nil {
		return "", err
	}
	u, err := s.storage.PresignGet(ctx, a.CVKey, 5*time.Minute, a.CVFilename)
	if err != nil {
		return "", err
	}
	if err := audit(ctx, s.store, actor, "cv_download", "application", a.ID, "Opened the CV of "+a.Name, nil, nil); err != nil {
		return "", err
	}
	return u, nil
}

func (s *Submissions) Outbox(ctx context.Context, actor Actor) ([]OutboxMail, error) {
	if !access.Can(access.Role(actor.Role), access.AuditRead) {
		return nil, apierror.New(apierror.CodeForbidden, "Only a superadmin can see outgoing mail.")
	}
	return s.store.ListOutbox(ctx, 200)
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// oneLine keeps a visitor's value from adding headers to a mail subject.
func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("\r", " ", "\n", " ").Replace(s)), " ")
}

func extOf(name string) string {
	if i := strings.LastIndex(name, "."); i > 0 {
		return name[i:]
	}
	return ""
}
