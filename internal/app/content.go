package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/certificate"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/contrast"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/params"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/placeholder"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/resource"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
)

func uuidOf(s string) uuid.UUID {
	u, _ := uuid.Parse(s)
	return u
}

// Content is the generic CMS use-case over every resource definition.
type Content struct {
	store    Store
	public   *Public
	settings *Settings
	now      func() time.Time
}

func NewContent(store Store, public *Public, settings *Settings, now func() time.Time) *Content {
	return &Content{store: store, public: public, settings: settings, now: now}
}

func (c *Content) def(name string) (resource.Def, error) {
	d, ok := resource.Lookup(name)
	if !ok {
		return resource.Def{}, apierror.NotFound("Section")
	}
	return d, nil
}

// Defs returns the definitions the role may see, with parameter groups the
// role may not edit removed from the group filter.
func (c *Content) Defs(role access.Role) []resource.Def {
	var out []resource.Def
	for _, d := range resource.Defs {
		if access.Can(role, access.ContentRead) || (d.Name == "parameters" && access.Can(role, access.ParamsContent)) {
			out = append(out, d)
		}
	}
	return out
}

func (c *Content) Options(ctx context.Context, role access.Role) (Options, error) {
	o, err := c.store.Options(ctx)
	if err != nil {
		return o, err
	}
	var groups []Option
	for _, g := range params.Groups {
		if !g.System || access.Can(role, access.ParamsSystem) {
			groups = append(groups, Option{Value: g.Key, Label: g.Label})
		}
	}
	o.Groups = groups
	return o, nil
}

func canRead(role access.Role, d resource.Def) bool {
	if d.Name == "parameters" {
		return access.Can(role, access.ParamsContent)
	}
	return access.Can(role, access.ContentRead)
}

func canWrite(role access.Role, d resource.Def) bool { return access.Can(role, d.WritePerm) }

// paramGroupAllowed — system parameter groups need params.system.
func paramGroupAllowed(role access.Role, grp string) bool {
	g, ok := params.GroupOf(grp)
	if !ok {
		return false
	}
	return !g.System || access.Can(role, access.ParamsSystem)
}

// List — BR-9.1: every list is searchable; the same query backs the export.
func (c *Content) List(ctx context.Context, actor Actor, name string, q ListQuery) ([]Row, int, error) {
	d, err := c.def(name)
	if err != nil {
		return nil, 0, err
	}
	role := access.Role(actor.Role)
	if !canRead(role, d) {
		return nil, 0, apierror.New(apierror.CodeForbidden, "You do not have access to this section.")
	}
	if q.Filters == nil {
		q.Filters = map[string]string{}
	}
	for k := range q.Filters {
		ok := false
		for _, f := range d.Filters {
			if f.Key == k {
				ok = true
			}
		}
		if !ok {
			return nil, 0, apierror.Validation("Unknown filter "+k, nil)
		}
	}
	rows, total, err := c.store.ListResource(ctx, d, q)
	if err != nil {
		return nil, 0, err
	}
	out := rows[:0]
	for _, r := range rows {
		if d.Name == "parameters" && !paramGroupAllowed(role, str0(r["grp"])) {
			total--
			continue
		}
		out = append(out, c.decorate(d, r))
	}
	return out, total, nil
}

func (c *Content) Get(ctx context.Context, actor Actor, name, rowID string) (Row, error) {
	d, err := c.def(name)
	if err != nil {
		return nil, err
	}
	role := access.Role(actor.Role)
	if !canRead(role, d) {
		return nil, apierror.New(apierror.CodeForbidden, "You do not have access to this section.")
	}
	r, err := c.store.GetResource(ctx, d, rowID, false)
	if errors.Is(err, ErrNotFound) {
		return nil, apierror.NotFound(capital(d.One))
	}
	if err != nil {
		return nil, err
	}
	if d.Name == "parameters" && !paramGroupAllowed(role, str0(r["grp"])) {
		return nil, apierror.NotFound("Parameter") // no existence oracle across roles
	}
	return c.decorate(d, r), nil
}

// decorate adds computed, read-only fields the CMS needs to explain itself:
// a certificate's status (BR-3.1), whether a parameter row can be deleted.
func (c *Content) decorate(d resource.Def, r Row) Row {
	switch d.Name {
	case "certificates":
		today := calendar.Today(c.now())
		var vu *calendar.Date
		if s := str0(r["valid_until"]); s != "" {
			if dt, err := calendar.ParseDate(s); err == nil {
				vu = &dt
			}
		}
		r["_status"] = string(certificate.StatusOn(vu, today))
	case "parameters":
		g, _ := params.GroupOf(str0(r["grp"]))
		r["_open"] = g.Open
		r["_system"] = g.System
		r["_groupLabel"] = g.Label
		if b, _ := r["is_secret"].(bool); b {
			r["value_en"], r["value_id"] = "••••••", "••••••"
		}
	}
	return r
}

func str0(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []string:
		return strings.Join(x, ",")
	}
	return fmt.Sprint(v)
}

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Create adds a row.
func (c *Content) Create(ctx context.Context, actor Actor, name string, raw map[string]any) (Row, error) {
	d, err := c.def(name)
	if err != nil {
		return nil, err
	}
	role := access.Role(actor.Role)
	if !canWrite(role, d) {
		return nil, apierror.New(apierror.CodeForbidden, "You cannot change this section.")
	}
	if !d.CanCreate {
		return nil, apierror.New(apierror.CodeConflict, d.CreateWhy)
	}
	vals, errs := resource.Validate(d, raw, true)
	if errs != nil {
		return nil, apierror.Validation("Please correct the highlighted fields.", errs)
	}
	if err := c.rules(ctx, role, d, nil, vals, true); err != nil {
		return nil, err
	}
	var newID string
	err = c.store.InTx(ctx, func(s Store) error {
		newID, err = s.InsertResource(ctx, d, vals, actor.UserID)
		if err != nil {
			return err
		}
		after, err := s.GetResource(ctx, d, newID, false)
		if err != nil {
			return err
		}
		return audit(ctx, s, actor, "create", d.Name, newID, "Created "+d.One+" "+label(d, after), nil, after)
	})
	if err != nil {
		return nil, err
	}
	c.invalidate(d)
	return c.Get(ctx, actor, name, newID)
}

// Update changes a row. The before-image is read under FOR UPDATE in the
// same transaction, so the placeholder rule and the audit diff see the row
// as it really was.
func (c *Content) Update(ctx context.Context, actor Actor, name, rowID string, raw map[string]any) (Row, error) {
	d, err := c.def(name)
	if err != nil {
		return nil, err
	}
	role := access.Role(actor.Role)
	if !canWrite(role, d) {
		return nil, apierror.New(apierror.CodeForbidden, "You cannot change this section.")
	}
	// The CMS posts the whole form, including fields that cannot change.
	// Unchanged immutable fields are not an attempt to change them.
	err = c.store.InTx(ctx, func(s Store) error {
		before, err := s.GetResource(ctx, d, rowID, true)
		if errors.Is(err, ErrNotFound) {
			return apierror.NotFound(capital(d.One))
		}
		if err != nil {
			return err
		}
		if d.Name == "parameters" && !paramGroupAllowed(role, str0(before["grp"])) {
			return apierror.NotFound("Parameter")
		}
		in := map[string]any{}
		for k, v := range raw {
			f, ok := d.Field(k)
			if !ok {
				continue // computed fields (_status, …) and unknown keys are ignored
			}
			if f.Immutable && sameValue(before[k], v) {
				continue
			}
			in[k] = v
		}
		vals, errs := resource.Validate(d, in, false)
		if errs != nil {
			return apierror.Validation("Please correct the highlighted fields.", errs)
		}
		if err := c.rules(ctx, role, d, before, vals, false); err != nil {
			return err
		}
		// BR-2.2 / BR-2.3.
		if d.Has("is_placeholder") {
			var explicit *bool
			if v, ok := vals["is_placeholder"].(bool); ok && v != boolOf(before["is_placeholder"]) {
				explicit = &v
			}
			changed := placeholder.ContentChanged(strMap(before), strMap(vals), d.ContentFields())
			vals["is_placeholder"] = placeholder.Next(boolOf(before["is_placeholder"]), changed, explicit)
		}
		diffBefore, diffAfter := diff(before, vals)
		if len(diffAfter) == 0 {
			return nil // nothing changed: no write, no audit noise
		}
		if err := s.UpdateResource(ctx, d, rowID, vals, actor.UserID); err != nil {
			return err
		}
		return audit(ctx, s, actor, "update", d.Name, rowID, "Updated "+d.One+" "+label(d, before), diffBefore, diffAfter)
	})
	if err != nil {
		return nil, err
	}
	c.invalidate(d)
	return c.Get(ctx, actor, name, rowID)
}

func (c *Content) Delete(ctx context.Context, actor Actor, name, rowID string) error {
	d, err := c.def(name)
	if err != nil {
		return err
	}
	role := access.Role(actor.Role)
	if !canWrite(role, d) {
		return apierror.New(apierror.CodeForbidden, "You cannot change this section.")
	}
	if !d.CanDelete {
		return apierror.New(apierror.CodeConflict, d.DeleteWhy)
	}
	err = c.store.InTx(ctx, func(s Store) error {
		before, err := s.GetResource(ctx, d, rowID, true)
		if errors.Is(err, ErrNotFound) {
			return apierror.NotFound(capital(d.One))
		}
		if err != nil {
			return err
		}
		if d.Name == "parameters" {
			grp := str0(before["grp"])
			if !paramGroupAllowed(role, grp) {
				return apierror.NotFound("Parameter")
			}
			if g, _ := params.GroupOf(grp); !g.Open {
				return apierror.New(apierror.CodeConflict, "This parameter is used by the site and cannot be deleted. You can change its value.")
			}
		}
		if err := s.DeleteResource(ctx, d, rowID); err != nil {
			return err
		}
		return audit(ctx, s, actor, "delete", d.Name, rowID, "Deleted "+d.One+" "+label(d, before), before, nil)
	})
	if err != nil {
		return err
	}
	c.invalidate(d)
	return nil
}

// rules are the resource-specific business rules on the merged row.
func (c *Content) rules(ctx context.Context, role access.Role, d resource.Def, before Row, vals map[string]any, creating bool) error {
	merged := Row{}
	for k, v := range before {
		merged[k] = v
	}
	for k, v := range vals {
		merged[k] = v
	}
	switch d.Name {
	case "certificates":
		// BR-3.2, with the date in the message so the editor knows what to fix.
		if boolOf(merged["published"]) {
			if s := str0(merged["valid_until"]); s != "" {
				dt, err := calendar.ParseDate(s)
				if err == nil && !certificate.CanPublish(&dt, calendar.Today(c.now())) {
					return apierror.Validation("This certificate expired and cannot be published.", map[string]string{
						"published": "Can't publish: this certificate expired on " + dt.Time().Format("2 Jan 2006") +
							". Upload the renewed certificate and update the date."})
				}
			}
		}
	case "services":
		// BR-11.1: text on the ink field and the label on white must be readable.
		errs := map[string]string{}
		if ok, r, err := contrast.MeetsAA(str0(merged["color_on_ink"]), str0(merged["color_ink"])); err == nil && !ok {
			errs["color_on_ink"] = fmt.Sprintf("Text on ink is %.2f:1 against the ink colour; it needs at least 4.5:1.", r)
		}
		if ok, r, err := contrast.MeetsAA(str0(merged["color_label"]), "#FFFFFF"); err == nil && !ok {
			errs["color_label"] = fmt.Sprintf("The label colour is %.2f:1 on white; it needs at least 4.5:1.", r)
		}
		if len(errs) > 0 {
			return apierror.Validation("These colours fail WCAG AA contrast.", errs)
		}
	case "parameters":
		grp := str0(merged["grp"])
		g, ok := params.GroupOf(grp)
		if !ok {
			return apierror.Validation("Choose an existing group.", map[string]string{"grp": "Unknown group."})
		}
		if !paramGroupAllowed(role, grp) {
			return apierror.New(apierror.CodeForbidden, "Only a superadmin can change "+g.Label+".")
		}
		if creating && !g.Open {
			return apierror.Validation("This group's keys are fixed by the site.", map[string]string{"grp": g.Label + " has a fixed set of keys; edit the existing ones."})
		}
		spec, ok := params.SpecFor(grp, str0(merged["key"]))
		if !ok {
			return apierror.Validation("Unknown parameter.", map[string]string{"key": "This key is not used by the site."})
		}
		en, err := params.Validate(spec, str0(merged["value_en"]))
		if err != nil {
			return apierror.Validation("The value is not valid.", map[string]string{"value_en": "Value " + err.Error() + "."})
		}
		vals["value_en"] = en
		if spec.Bilingual {
			idv, err := params.Validate(spec, str0(merged["value_id"]))
			if err != nil {
				return apierror.Validation("The value is not valid.", map[string]string{"value_id": "Value " + err.Error() + "."})
			}
			if strings.TrimSpace(idv) == "" && strings.TrimSpace(en) != "" {
				return apierror.Validation("Both languages are needed.", map[string]string{"value_id": "Enter the Indonesian value too."})
			}
			vals["value_id"] = idv
		} else {
			vals["value_id"] = en
		}
	}
	return nil
}

func (c *Content) invalidate(d resource.Def) {
	c.public.Invalidate()
	if d.Name == "parameters" {
		c.settings.Invalidate()
	}
}

func label(d resource.Def, r Row) string {
	if v := str0(r[d.NameField]); v != "" {
		if len([]rune(v)) > 60 {
			v = string([]rune(v)[:60]) + "…"
		}
		return "“" + v + "”"
	}
	return ""
}

func boolOf(v any) bool { b, _ := v.(bool); return b }

func sameValue(a, b any) bool { return str0(a) == str0(b) }

func strMap(r map[string]any) map[string]string {
	out := make(map[string]string, len(r))
	for k, v := range r {
		out[k] = str0(v)
	}
	return out
}

// diff returns only the changed fields, before and after.
func diff(before Row, vals map[string]any) (map[string]any, map[string]any) {
	b, a := map[string]any{}, map[string]any{}
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if str0(before[k]) != str0(vals[k]) {
			b[k], a[k] = before[k], vals[k]
		}
	}
	return b, a
}

// Placeholders returns the remaining placeholder count per resource the
// role can read.
func (c *Content) Placeholders(ctx context.Context, actor Actor) (map[string]int, error) {
	all, err := c.store.PlaceholderCounts(ctx)
	if err != nil {
		return nil, err
	}
	role := access.Role(actor.Role)
	out := map[string]int{}
	for name, n := range all {
		if d, ok := resource.Lookup(name); ok && canRead(role, d) {
			out[name] = n
		}
	}
	return out, nil
}
