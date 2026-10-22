package app

import (
	"context"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
)

// Actor is who is doing something, carried from the HTTP edge into every
// use-case so the audit row can say who, from where and in which request.
type Actor struct {
	UserID    string
	Email     string
	Role      string
	IP        string
	UserAgent string
	RequestID string
}

// System is the actor for scheduled jobs.
func System(job string) Actor { return Actor{Email: "system:" + job} }

func (a Actor) entry(action, entity, entityID, summary string, before, after any) AuditEntry {
	return AuditEntry{
		ID: id.NewString(), At: time.Now().UTC(), ActorID: a.UserID, ActorEmail: a.Email,
		Action: action, Entity: entity, EntityID: entityID, Summary: summary,
		Before: before, After: after, IP: a.IP, UserAgent: a.UserAgent, RequestID: a.RequestID,
	}
}

// audit writes the row in the same transaction as the change (BR-10.1): a
// change that happened without its audit row did not happen.
func audit(ctx context.Context, s Store, a Actor, action, entity, entityID, summary string, before, after any) error {
	return s.AppendAudit(ctx, a.entry(action, entity, entityID, summary, before, after))
}

// AuditLog is the read side of the audit trail (superadmin only).
type AuditLog struct{ store Store }

func NewAuditLog(store Store) *AuditLog { return &AuditLog{store: store} }

func (a *AuditLog) List(ctx context.Context, actor Actor, q ListQuery) ([]AuditEntry, int, error) {
	if !access.Can(access.Role(actor.Role), access.AuditRead) {
		return nil, 0, apierror.New(apierror.CodeForbidden, "Only a superadmin can read the audit history.")
	}
	return a.store.ListAudit(ctx, q)
}
