package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/params"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/resource"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
)

func listQuery(c *gin.Context) app.ListQuery {
	q := app.ListQuery{Q: c.Query("q"), Filters: map[string]string{}}
	for k, v := range c.Request.URL.Query() {
		if strings.HasPrefix(k, "f.") && len(v) > 0 && v[0] != "" {
			q.Filters[strings.TrimPrefix(k, "f.")] = v[0]
		}
	}
	q.Limit, _ = strconv.Atoi(c.Query("limit"))
	q.Offset, _ = strconv.Atoi(c.Query("offset"))
	return q
}

// getSchema is the single source the CMS builds its screens from.
func (s *Server) getSchema(c *gin.Context) {
	a := actorOf(c)
	type groupV struct {
		Key    string `json:"key"`
		Label  string `json:"label"`
		System bool   `json:"system"`
		Open   bool   `json:"open"`
	}
	var groups []groupV
	for _, g := range params.Groups {
		if !g.System || access.Can(access.Role(a.Role), access.ParamsSystem) {
			groups = append(groups, groupV{g.Key, g.Label, g.System, g.Open})
		}
	}
	ok(c, gin.H{"resources": s.Content.Defs(access.Role(a.Role)), "icons": resource.Icons, "paramGroups": groups})
}

func (s *Server) getOptions(c *gin.Context) {
	o, err := s.Content.Options(c.Request.Context(), access.Role(actorOf(c).Role))
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, o)
}

func (s *Server) getPlaceholders(c *gin.Context) {
	m, err := s.Content.Placeholders(c.Request.Context(), actorOf(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, m)
}

func (s *Server) listResource(c *gin.Context) {
	name := c.Param("name")
	rows, total, err := s.Content.List(c.Request.Context(), actorOf(c), name, listQuery(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	if wantsCSV(c) {
		s.resourceCSV(c, name, rows)
		return
	}
	ok(c, gin.H{"rows": rows, "total": total})
}

// resourceCSV exports exactly the rows on screen (same search and filters),
// with references shown by name rather than by id (BR-9.2).
func (s *Server) resourceCSV(c *gin.Context, name string, rows []app.Row) {
	d, _ := resource.Lookup(name)
	opts, err := s.Content.Options(c.Request.Context(), access.Role(actorOf(c).Role))
	if err != nil {
		s.fail(c, err)
		return
	}
	media, err := s.Media.List(c.Request.Context(), actorOf(c), "", "")
	if err != nil {
		s.fail(c, err)
		return
	}
	mediaName := map[string]string{}
	for _, m := range media {
		mediaName[m.ID] = m.Name
	}
	refName := func(ref, v string) string {
		for _, o := range opts.Refs[ref] {
			if o.Value == v {
				return o.Label
			}
		}
		return v
	}
	header := []string{}
	for _, f := range d.Fields {
		header = append(header, f.Label)
	}
	if name == "certificates" {
		header = append(header, "Status")
	}
	header = append(header, "Updated")
	var out [][]string
	for _, r := range rows {
		var line []string
		for _, f := range d.Fields {
			v := r[f.Key]
			switch f.Type {
			case resource.Bool:
				if b, _ := v.(bool); b {
					line = append(line, "Yes")
				} else {
					line = append(line, "No")
				}
			case resource.Media:
				line = append(line, mediaName[fmt.Sprint(nz(v))])
			case resource.Ref:
				line = append(line, refName(f.Ref, fmt.Sprint(nz(v))))
			case resource.Refs:
				var names []string
				if ids, ok := v.([]string); ok {
					for _, x := range ids {
						names = append(names, refName(f.Ref, x))
					}
				}
				line = append(line, strings.Join(names, ", "))
			default:
				line = append(line, fmt.Sprint(nz(v)))
			}
		}
		if name == "certificates" {
			line = append(line, fmt.Sprint(nz(r["_status"])))
		}
		line = append(line, fmt.Sprint(nz(r["updated_at"])))
		out = append(out, line)
	}
	writeCSV(c, name, header, out)
}

func nz(v any) any {
	if v == nil {
		return ""
	}
	return v
}

func (s *Server) getResource(c *gin.Context) {
	r, err := s.Content.Get(c.Request.Context(), actorOf(c), c.Param("name"), c.Param("id"))
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, r)
}

func bindMap(c *gin.Context) (map[string]any, error) {
	var m map[string]any
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&m); err != nil {
		return nil, apierror.Validation("The request is not valid JSON.", nil)
	}
	return m, nil
}

func (s *Server) createResource(c *gin.Context) {
	m, err := bindMap(c)
	if err != nil {
		s.fail(c, err)
		return
	}
	r, err := s.Content.Create(c.Request.Context(), actorOf(c), c.Param("name"), m)
	if err != nil {
		s.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (s *Server) updateResource(c *gin.Context) {
	m, err := bindMap(c)
	if err != nil {
		s.fail(c, err)
		return
	}
	r, err := s.Content.Update(c.Request.Context(), actorOf(c), c.Param("name"), c.Param("id"), m)
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, r)
}

func (s *Server) deleteResource(c *gin.Context) {
	if err := s.Content.Delete(c.Request.Context(), actorOf(c), c.Param("name"), c.Param("id")); err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

// ---------------------------------------------------------------------------
// Media
// ---------------------------------------------------------------------------

type mediaV struct {
	app.MediaRecord
	Thumb string `json:"thumb"`
	URL   string `json:"url"`
}

func view(m app.MediaRecord) mediaV {
	v := mediaV{MediaRecord: m}
	if m.Hash == "" {
		return v
	}
	if m.Mime == "image/svg+xml" {
		v.Thumb, v.URL = app.ImageURL(m, 0), app.ImageURL(m, 0)
		return v
	}
	min, max := 0, 0
	for _, x := range m.Variants {
		if min == 0 || x.W < min {
			min = x.W
		}
		if x.W > max {
			max = x.W
		}
	}
	if min > 0 {
		v.Thumb, v.URL = app.ImageURL(m, min), app.ImageURL(m, max)
	}
	return v
}

func (s *Server) listMedia(c *gin.Context) {
	list, err := s.Media.List(c.Request.Context(), actorOf(c), c.Query("q"), c.Query("kind"))
	if err != nil {
		s.fail(c, err)
		return
	}
	if wantsCSV(c) {
		var rows [][]string
		for _, m := range list {
			rows = append(rows, []string{m.Name, m.Kind, m.Mime, strconv.Itoa(m.Width) + "×" + strconv.Itoa(m.Height),
				strconv.FormatInt(m.Bytes, 10), strconv.Itoa(m.Usage), view(m).URL, ts(m.CreatedAt)})
		}
		writeCSV(c, "media", []string{"Name", "Type", "Format", "Size (px)", "Bytes", "Used by", "URL", "Added"}, rows)
		return
	}
	out := make([]mediaV, 0, len(list))
	for _, m := range list {
		out = append(out, view(m))
	}
	ok(c, gin.H{"rows": out, "total": len(out)})
}

func (s *Server) uploadMedia(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		s.fail(c, apierror.Validation("Choose an image to upload.", map[string]string{"file": "Choose an image to upload."}))
		return
	}
	f, err := fh.Open()
	if err != nil {
		s.fail(c, err)
		return
	}
	data, err := io.ReadAll(io.LimitReader(f, 51<<20))
	_ = f.Close()
	if err != nil {
		s.fail(c, err)
		return
	}
	m, err := s.Media.Upload(c.Request.Context(), actorOf(c), fh.Filename, c.PostForm("kind"), data)
	if err != nil {
		s.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, view(*m))
}

func (s *Server) updateMedia(c *gin.Context) {
	var b struct{ Name, AltEn, AltId string }
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Invalid request.", nil))
		return
	}
	m, err := s.Media.UpdateMeta(c.Request.Context(), actorOf(c), c.Param("id"), b.Name, b.AltEn, b.AltId)
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, view(*m))
}

func (s *Server) deleteMedia(c *gin.Context) {
	if err := s.Media.Delete(c.Request.Context(), actorOf(c), c.Param("id")); err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

// ---------------------------------------------------------------------------
// Inbox
// ---------------------------------------------------------------------------

func (s *Server) listQuotations(c *gin.Context) {
	list, total, err := s.Submissions.Quotations(c.Request.Context(), actorOf(c), listQuery(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	if wantsCSV(c) {
		var rows [][]string
		for _, q := range list {
			area := ""
			if q.AreaM2 != nil {
				area = strconv.Itoa(*q.AreaM2)
			}
			rows = append(rows, []string{ts(q.CreatedAt), q.Name, q.Company, q.Email, q.Phone, q.BuildingType, area,
				q.Location, strings.Join(q.Services, ", "), q.Message, strings.ToUpper(q.Lang)})
		}
		writeCSV(c, "quotations", []string{"Received", "Name", "Company", "Email", "Phone", "Building type", "Area (m²)",
			"Location", "Services", "Message", "Language"}, rows)
		return
	}
	ok(c, gin.H{"rows": list, "total": total})
}

func (s *Server) getQuotation(c *gin.Context) {
	q, err := s.Submissions.Quotation(c.Request.Context(), actorOf(c), c.Param("id"))
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, q)
}

func (s *Server) listApplications(c *gin.Context) {
	list, total, err := s.Submissions.Applications(c.Request.Context(), actorOf(c), listQuery(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	if wantsCSV(c) {
		var rows [][]string
		for _, a := range list {
			rows = append(rows, []string{ts(a.CreatedAt), a.Name, a.Phone, a.Email, a.Position, a.CVFilename, a.Message, strings.ToUpper(a.Lang)})
		}
		writeCSV(c, "applications", []string{"Received", "Name", "Phone", "Email", "Position", "CV file", "Message", "Language"}, rows)
		return
	}
	ok(c, gin.H{"rows": list, "total": total})
}

func (s *Server) getApplication(c *gin.Context) {
	a, err := s.Submissions.Application(c.Request.Context(), actorOf(c), c.Param("id"))
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, a)
}

func (s *Server) cvLink(c *gin.Context) {
	u, err := s.Submissions.CVLink(c.Request.Context(), actorOf(c), c.Param("id"))
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"url": u})
}

// ---------------------------------------------------------------------------
// Users, audit, outbox
// ---------------------------------------------------------------------------

func (s *Server) listUsers(c *gin.Context) {
	list, err := s.auth.ListUsers(c.Request.Context(), c.Query("q"))
	if err != nil {
		s.fail(c, err)
		return
	}
	if wantsCSV(c) {
		var rows [][]string
		for _, u := range list {
			last := ""
			if u.LastLoginAt != nil {
				last = ts(*u.LastLoginAt)
			}
			active := "Yes"
			if !u.Active {
				active = "No"
			}
			rows = append(rows, []string{u.Name, u.Email, u.Role, active, last, ts(u.CreatedAt)})
		}
		writeCSV(c, "users", []string{"Name", "Email", "Role", "Active", "Last sign-in", "Created"}, rows)
		return
	}
	ok(c, gin.H{"rows": list, "total": len(list)})
}

func (s *Server) createUser(c *gin.Context) {
	var b struct{ Email, Name, Role, Password string }
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Invalid request.", nil))
		return
	}
	u, err := s.auth.CreateUser(c.Request.Context(), actorOf(c), b.Email, b.Name, b.Role, b.Password)
	if err != nil {
		s.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (s *Server) updateUser(c *gin.Context) {
	var b struct {
		Name, Role string
		Active     bool
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Invalid request.", nil))
		return
	}
	u, err := s.auth.UpdateUser(c.Request.Context(), actorOf(c), c.Param("id"), b.Name, b.Role, b.Active)
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, u)
}

func (s *Server) resetPassword(c *gin.Context) {
	var b struct{ Password string }
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Invalid request.", nil))
		return
	}
	if err := s.auth.ResetPassword(c.Request.Context(), actorOf(c), c.Param("id"), b.Password); err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (s *Server) listAudit(c *gin.Context) {
	list, total, err := s.Audit.List(c.Request.Context(), actorOf(c), listQuery(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	if wantsCSV(c) {
		var rows [][]string
		for _, e := range list {
			before, _ := json.Marshal(e.Before)
			after, _ := json.Marshal(e.After)
			rows = append(rows, []string{ts(e.At), e.ActorEmail, e.Action, e.Entity, e.EntityID, e.Summary,
				string(before), string(after), e.IP, e.RequestID})
		}
		writeCSV(c, "audit", []string{"When", "Who", "Action", "Entity", "Entity id", "Summary", "Before", "After", "IP", "Request id"}, rows)
		return
	}
	ok(c, gin.H{"rows": list, "total": total})
}

func (s *Server) listOutbox(c *gin.Context) {
	list, err := s.Submissions.Outbox(c.Request.Context(), actorOf(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))
	var out []app.OutboxMail
	for _, m := range list {
		if q == "" || strings.Contains(strings.ToLower(m.Subject+" "+m.Kind+" "+strings.Join(m.Recipients, " ")+" "+m.LastError), q) {
			out = append(out, m)
		}
	}
	if out == nil {
		out = []app.OutboxMail{}
	}
	if wantsCSV(c) {
		var rows [][]string
		for _, m := range out {
			sent := ""
			if m.SentAt != nil {
				sent = ts(*m.SentAt)
			}
			rows = append(rows, []string{ts(m.CreatedAt), m.Kind, m.Subject, strings.Join(m.Recipients, ", "), strconv.Itoa(m.Attempts), sent, m.LastError})
		}
		writeCSV(c, "outbox", []string{"Created", "Kind", "Subject", "To", "Attempts", "Sent", "Last error"}, rows)
		return
	}
	ok(c, gin.H{"rows": out, "total": len(out)})
}
