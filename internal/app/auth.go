package app

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/sanitize"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/security"
)

// ErrNotFound is what a repository returns for a missing row.
var ErrNotFound = errors.New("not found")

// The one message for every failed sign-in (BR-8.3). Wrong password,
// unknown email, locked and deactivated accounts all read the same, so the
// login form cannot be used to find out who works here.
const badLogin = "Email or password is incorrect."

type Auth struct {
	store      Store
	settings   *Settings
	tokens     *security.TokenIssuer
	refreshTTL time.Duration
	setupToken string
	now        func() time.Time
}

func NewAuth(store Store, settings *Settings, tokens *security.TokenIssuer, refreshTTL time.Duration, setupToken string, now func() time.Time) *Auth {
	return &Auth{store: store, settings: settings, tokens: tokens, refreshTTL: refreshTTL, setupToken: setupToken, now: now}
}

type Session struct {
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"-"`
	RefreshExpiresAt time.Time `json:"-"`
	User             *Me       `json:"user"`
}

type Me struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func meOf(u *User) *Me {
	r, _ := access.ParseRole(u.Role)
	var perms []string
	for _, p := range access.Permissions(r) {
		perms = append(perms, string(p))
	}
	return &Me{ID: u.ID, Email: u.Email, Name: u.Name, Role: u.Role, Permissions: perms}
}

// SetupNeeded reports whether the first-run screen should show.
func (a *Auth) SetupNeeded(ctx context.Context) (bool, error) {
	n, err := a.store.CountUsers(ctx)
	return n == 0, err
}

// Setup creates the first superadmin (BR-8.2). It works once: when any user
// exists it is closed, and it needs the setup token from the server's
// configuration, so a fresh install reachable from the internet cannot be
// claimed by whoever finds it first.
func (a *Auth) Setup(ctx context.Context, token, email, name, password string, actor Actor) (*Session, error) {
	if a.setupToken == "" {
		return nil, apierror.New(apierror.CodeSetupClosed, "First-run setup is not enabled on this server.")
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(a.setupToken)) != 1 {
		return nil, apierror.Validation("The setup token is not correct.", map[string]string{"token": "The setup token is not correct."})
	}
	var created *User
	err := a.store.InTx(ctx, func(s Store) error {
		n, err := s.CountUsers(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return apierror.New(apierror.CodeSetupClosed, "Setup has already been completed.")
		}
		u, err := a.newUser(ctx, email, name, string(access.RoleSuperadmin), password)
		if err != nil {
			return err
		}
		if err := s.CreateUser(ctx, *u); err != nil {
			return err
		}
		created = u
		actor.UserID, actor.Email = u.ID, u.Email
		return audit(ctx, s, actor, "setup", "user", u.ID, "First superadmin created", nil, map[string]any{"email": u.Email, "role": u.Role})
	})
	if err != nil {
		return nil, err
	}
	return a.issue(ctx, created, actor.IP, actor.UserAgent)
}

func (a *Auth) newUser(ctx context.Context, email, name, role, password string) (*User, error) {
	errs := map[string]string{}
	em, err := sanitize.Email(email)
	if err != nil {
		errs["email"] = "Enter a valid email address."
	}
	nm, err := sanitize.Text(name, 120)
	if err != nil {
		errs["name"] = "Enter a name (up to 120 characters)."
	}
	if _, ok := access.ParseRole(role); !ok {
		errs["role"] = "Choose a role."
	}
	if msg := a.passwordProblem(ctx, password); msg != "" {
		errs["password"] = msg
	}
	if len(errs) > 0 {
		return nil, apierror.Validation("Please correct the highlighted fields.", errs)
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := a.now().UTC()
	return &User{ID: id.NewString(), Email: strings.ToLower(em), Name: nm, PasswordHash: hash, Role: role,
		Active: true, PasswordChangedAt: now, CreatedAt: now}, nil
}

func (a *Auth) passwordProblem(ctx context.Context, pw string) string {
	min := a.settings.Int(ctx, "auth.password_min_length", security.DefaultPasswordMinLength)
	if err := security.CheckPasswordStrength(pw, min); err != nil {
		return "Use at least " + itoa(min) + " characters."
	}
	if len(pw) > 256 {
		return "Use at most 256 characters."
	}
	return ""
}

// Login — BR-8.3.
func (a *Auth) Login(ctx context.Context, email, password, ip, ua string) (*Session, error) {
	em := strings.ToLower(strings.TrimSpace(email))
	u, err := a.store.UserByEmail(ctx, em)
	if errors.Is(err, ErrNotFound) {
		security.SpendDummyVerify(password) // same cost as a real check
		return nil, apierror.New(apierror.CodeUnauthenticated, badLogin)
	}
	if err != nil {
		return nil, err
	}
	now := a.now()
	if !u.Active || (u.LockedUntil != nil && u.LockedUntil.After(now)) {
		security.SpendDummyVerify(password)
		return nil, apierror.New(apierror.CodeUnauthenticated, badLogin)
	}
	if err := security.VerifyPassword(password, u.PasswordHash); err != nil {
		var lock *time.Time
		max := a.settings.Int(ctx, "auth.max_failed_logins", 5)
		if u.FailedLogins+1 >= max {
			t := now.Add(time.Duration(a.settings.Int(ctx, "auth.lockout_minutes", 15)) * time.Minute)
			lock = &t
		}
		if err := a.store.RecordLoginFailure(ctx, u.ID, lock); err != nil {
			return nil, err
		}
		if lock != nil {
			_ = audit(ctx, a.store, Actor{UserID: u.ID, Email: u.Email, IP: ip, UserAgent: ua}, "lockout", "user", u.ID,
				"Account locked after repeated failed sign-ins", nil, map[string]any{"lockedUntil": lock})
		}
		return nil, apierror.New(apierror.CodeUnauthenticated, badLogin)
	}
	if err := a.store.RecordLoginSuccess(ctx, u.ID); err != nil {
		return nil, err
	}
	return a.issue(ctx, u, ip, ua)
}

func (a *Auth) issue(ctx context.Context, u *User, ip, ua string) (*Session, error) {
	now := a.now()
	access, _, exp, err := a.tokens.Issue(uuidOf(u.ID), now)
	if err != nil {
		return nil, err
	}
	raw, err := security.NewToken()
	if err != nil {
		return nil, err
	}
	rt := RefreshToken{ID: id.NewString(), UserID: u.ID, TokenHash: security.HashToken(raw),
		ExpiresAt: now.Add(a.refreshTTL), IP: ip, UserAgent: ua}
	if err := a.store.CreateRefreshToken(ctx, rt); err != nil {
		return nil, err
	}
	return &Session{AccessToken: access, AccessExpiresAt: exp, RefreshToken: raw, RefreshExpiresAt: rt.ExpiresAt, User: meOf(u)}, nil
}

// Refresh rotates the refresh token. Presenting a token that was already
// rotated means it was copied: every session of that user is revoked.
func (a *Auth) Refresh(ctx context.Context, raw, ip, ua string) (*Session, error) {
	if raw == "" {
		return nil, apierror.New(apierror.CodeUnauthenticated, "Please sign in.")
	}
	var sess *Session
	err := a.store.InTx(ctx, func(s Store) error {
		rt, err := s.RefreshTokenByHash(ctx, security.HashToken(raw))
		if errors.Is(err, ErrNotFound) {
			return apierror.New(apierror.CodeUnauthenticated, "Please sign in.")
		}
		if err != nil {
			return err
		}
		if rt.RevokedAt != nil {
			_ = s.RevokeUserRefreshTokens(ctx, rt.UserID)
			return apierror.New(apierror.CodeUnauthenticated, "Please sign in.")
		}
		if a.now().After(rt.ExpiresAt) {
			return apierror.New(apierror.CodeUnauthenticated, "Please sign in.")
		}
		u, err := s.UserByID(ctx, rt.UserID)
		if err != nil || !u.Active {
			return apierror.New(apierror.CodeUnauthenticated, "Please sign in.")
		}
		inner := &Auth{store: s, settings: a.settings, tokens: a.tokens, refreshTTL: a.refreshTTL, now: a.now}
		sess, err = inner.issue(ctx, u, ip, ua)
		if err != nil {
			return err
		}
		newRT, err := s.RefreshTokenByHash(ctx, security.HashToken(sess.RefreshToken))
		if err != nil {
			return err
		}
		return s.RevokeRefreshToken(ctx, rt.ID, newRT.ID)
	})
	return sess, err
}

func (a *Auth) Logout(ctx context.Context, jti string, accessExp time.Time, refreshRaw string) error {
	if jti != "" {
		if err := a.store.RevokeJTI(ctx, jti, accessExp); err != nil {
			return err
		}
	}
	if refreshRaw != "" {
		rt, err := a.store.RefreshTokenByHash(ctx, security.HashToken(refreshRaw))
		if err == nil {
			return a.store.RevokeRefreshToken(ctx, rt.ID, "")
		}
	}
	return nil
}

// Authenticate resolves an access token to the current user. Permissions
// come from the database on every request, so a role change or a
// deactivation takes effect immediately rather than in fifteen minutes.
func (a *Auth) Authenticate(ctx context.Context, token string) (*User, *security.Claims, error) {
	c, err := a.tokens.Parse(token)
	if err != nil {
		return nil, nil, apierror.New(apierror.CodeUnauthenticated, "Your session has expired. Please sign in again.")
	}
	revoked, err := a.store.IsJTIRevoked(ctx, c.JTI)
	if err != nil {
		return nil, nil, err
	}
	if revoked {
		return nil, nil, apierror.New(apierror.CodeUnauthenticated, "Your session has ended. Please sign in again.")
	}
	u, err := a.store.UserByID(ctx, c.UserID.String())
	if err != nil || !u.Active {
		return nil, nil, apierror.New(apierror.CodeUnauthenticated, "Please sign in.")
	}
	// A password change ends every session issued before it.
	if c.IssuedAt != nil && c.IssuedAt.Time.Add(time.Second).Before(u.PasswordChangedAt) {
		return nil, nil, apierror.New(apierror.CodeUnauthenticated, "Your password was changed. Please sign in again.")
	}
	return u, c, nil
}

func (a *Auth) Me(u *User) *Me { return meOf(u) }

func (a *Auth) ChangePassword(ctx context.Context, actor Actor, current, next string) error {
	u, err := a.store.UserByID(ctx, actor.UserID)
	if err != nil {
		return err
	}
	if security.VerifyPassword(current, u.PasswordHash) != nil {
		return apierror.Validation("The current password is not correct.", map[string]string{"current": "The current password is not correct."})
	}
	if msg := a.passwordProblem(ctx, next); msg != "" {
		return apierror.Validation(msg, map[string]string{"password": msg})
	}
	hash, err := security.HashPassword(next)
	if err != nil {
		return err
	}
	return a.store.InTx(ctx, func(s Store) error {
		if err := s.SetPassword(ctx, u.ID, hash); err != nil {
			return err
		}
		if err := s.RevokeUserRefreshTokens(ctx, u.ID); err != nil {
			return err
		}
		return audit(ctx, s, actor, "password_change", "user", u.ID, "Changed own password", nil, nil)
	})
}

// ---------------------------------------------------------------------------
// User management (superadmin)
// ---------------------------------------------------------------------------

func (a *Auth) ListUsers(ctx context.Context, q string) ([]User, error) {
	return a.store.ListUsers(ctx, q)
}

func (a *Auth) CreateUser(ctx context.Context, actor Actor, email, name, role, password string) (*User, error) {
	u, err := a.newUser(ctx, email, name, role, password)
	if err != nil {
		return nil, err
	}
	err = a.store.InTx(ctx, func(s Store) error {
		if _, err := s.UserByEmail(ctx, u.Email); err == nil {
			return apierror.Validation("This email already has an account.", map[string]string{"email": "This email already has an account."})
		}
		if err := s.CreateUser(ctx, *u); err != nil {
			return err
		}
		return audit(ctx, s, actor, "create", "user", u.ID, "Created "+u.Email, nil, map[string]any{"email": u.Email, "name": u.Name, "role": u.Role})
	})
	return u, err
}

// UpdateUser — BR-8.4: nobody can demote or deactivate themselves, and the
// last active superadmin cannot be demoted or deactivated by anyone. Without
// it, one click locks the company out of its own CMS.
func (a *Auth) UpdateUser(ctx context.Context, actor Actor, userID, name, role string, active bool) (*User, error) {
	var out *User
	err := a.store.InTx(ctx, func(s Store) error {
		u, err := s.UserByID(ctx, userID)
		if errors.Is(err, ErrNotFound) {
			return apierror.NotFound("User")
		}
		if err != nil {
			return err
		}
		before := map[string]any{"name": u.Name, "role": u.Role, "active": u.Active}
		if _, ok := access.ParseRole(role); !ok {
			return apierror.Validation("Choose a role.", map[string]string{"role": "Choose a role."})
		}
		nm, err := sanitize.Text(name, 120)
		if err != nil {
			return apierror.Validation("Enter a name.", map[string]string{"name": "Enter a name (up to 120 characters)."})
		}
		losing := u.Role == string(access.RoleSuperadmin) && u.Active && (role != string(access.RoleSuperadmin) || !active)
		if losing && u.ID == actor.UserID {
			return apierror.New(apierror.CodeConflict, "You cannot remove your own superadmin access. Ask another superadmin.")
		}
		if losing {
			n, err := s.CountActiveSuperadmins(ctx)
			if err != nil {
				return err
			}
			if n <= 1 {
				return apierror.New(apierror.CodeConflict, "This is the last active superadmin. Make someone else superadmin first.")
			}
		}
		u.Name, u.Role, u.Active = nm, role, active
		if err := s.UpdateUser(ctx, *u); err != nil {
			return err
		}
		if !active {
			if err := s.RevokeUserRefreshTokens(ctx, u.ID); err != nil {
				return err
			}
		}
		out = u
		return audit(ctx, s, actor, "update", "user", u.ID, "Updated "+u.Email, before, map[string]any{"name": u.Name, "role": u.Role, "active": u.Active})
	})
	return out, err
}

// ResetPassword sets a new password for another user and ends their
// sessions. There is no self-service reset in phase 1 (decision Q12).
func (a *Auth) ResetPassword(ctx context.Context, actor Actor, userID, password string) error {
	if msg := a.passwordProblem(ctx, password); msg != "" {
		return apierror.Validation(msg, map[string]string{"password": msg})
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	return a.store.InTx(ctx, func(s Store) error {
		u, err := s.UserByID(ctx, userID)
		if errors.Is(err, ErrNotFound) {
			return apierror.NotFound("User")
		}
		if err != nil {
			return err
		}
		if err := s.SetPassword(ctx, u.ID, hash); err != nil {
			return err
		}
		if err := s.RevokeUserRefreshTokens(ctx, u.ID); err != nil {
			return err
		}
		return audit(ctx, s, actor, "password_reset", "user", u.ID, "Reset the password of "+u.Email, nil, nil)
	})
}

// SeedDevAdmin creates the development superadmin. It refuses to run in
// production (CLAUDE.md §7: no default admin credentials) — the caller
// passes the environment, and the command line does too.
func (a *Auth) SeedDevAdmin(ctx context.Context, env, email, password string) (bool, error) {
	if env == "production" {
		return false, errors.New("refusing to seed a development admin in production")
	}
	if _, err := a.store.UserByEmail(ctx, email); err == nil {
		return false, nil
	}
	u, err := a.newUser(ctx, email, "Development Admin", string(access.RoleSuperadmin), password)
	if err != nil {
		return false, err
	}
	err = a.store.InTx(ctx, func(s Store) error {
		if err := s.CreateUser(ctx, *u); err != nil {
			return err
		}
		return audit(ctx, s, System("seed-dev"), "create", "user", u.ID, "Development superadmin seeded", nil, map[string]any{"email": u.Email})
	})
	return err == nil, err
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// BreakGlassReset sets a new password for an account from the server's
// command line and unlocks and reactivates it — the way back in when every
// superadmin is locked out. It is audited as a system action.
func (a *Auth) BreakGlassReset(ctx context.Context, email, password string) error {
	u, err := a.store.UserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, ErrNotFound) {
		return fmt.Errorf("no account with that email")
	}
	if err != nil {
		return err
	}
	if msg := a.passwordProblem(ctx, password); msg != "" {
		return errors.New(msg)
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	return a.store.InTx(ctx, func(s Store) error {
		if err := s.SetPassword(ctx, u.ID, hash); err != nil {
			return err
		}
		u.Active = true
		if err := s.UpdateUser(ctx, *u); err != nil {
			return err
		}
		if err := s.RevokeUserRefreshTokens(ctx, u.ID); err != nil {
			return err
		}
		return audit(ctx, s, System("reset-password"), "password_reset", "user", u.ID, "Break-glass reset from the server console for "+u.Email, nil, nil)
	})
}
