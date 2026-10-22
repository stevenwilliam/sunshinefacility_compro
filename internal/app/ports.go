// Package app holds the use-cases. It orchestrates the domain rules and the
// ports below; it knows nothing about HTTP, SQL drivers, S3 or SMTP.
package app

import (
	"context"
	"io"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/resource"
)

// Row is one record of a CMS resource, keyed by column.
type Row = map[string]any

// ListQuery is a search over a resource. Filters are column → value for
// columns the definition declares as filters; the repository refuses others.
type ListQuery struct {
	Q       string
	Filters map[string]string
	Limit   int
	Offset  int
}

// Store is the persistence port. The postgres adapter implements it.
type Store interface {
	InTx(ctx context.Context, fn func(Store) error) error

	// Users and sessions.
	CountUsers(ctx context.Context) (int, error)
	CountActiveSuperadmins(ctx context.Context) (int, error)
	UserByEmail(ctx context.Context, email string) (*User, error)
	UserByID(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context, q string) ([]User, error)
	CreateUser(ctx context.Context, u User) error
	UpdateUser(ctx context.Context, u User) error
	SetPassword(ctx context.Context, userID, hash string) error
	RecordLoginFailure(ctx context.Context, userID string, lockUntil *time.Time) error
	RecordLoginSuccess(ctx context.Context, userID string) error
	CreateRefreshToken(ctx context.Context, t RefreshToken) error
	RefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string, replacedBy string) error
	RevokeUserRefreshTokens(ctx context.Context, userID string) error
	RevokeJTI(ctx context.Context, jti string, until time.Time) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)

	// Audit (append-only).
	AppendAudit(ctx context.Context, e AuditEntry) error
	ListAudit(ctx context.Context, q ListQuery) ([]AuditEntry, int, error)

	// Generic CMS resources.
	ListResource(ctx context.Context, d resource.Def, q ListQuery) ([]Row, int, error)
	GetResource(ctx context.Context, d resource.Def, id string, forUpdate bool) (Row, error)
	InsertResource(ctx context.Context, d resource.Def, vals map[string]any, actorID string) (string, error)
	UpdateResource(ctx context.Context, d resource.Def, id string, vals map[string]any, actorID string) error
	DeleteResource(ctx context.Context, d resource.Def, id string) error
	Options(ctx context.Context) (Options, error)
	// PlaceholderCounts is how many rows of each resource are still placeholders.
	PlaceholderCounts(ctx context.Context) (map[string]int, error)

	// Parameters.
	AllParameters(ctx context.Context) ([]Parameter, error)

	// Media.
	InsertMedia(ctx context.Context, m MediaRecord) error
	UpdateMediaFile(ctx context.Context, m MediaRecord) error
	UpdateMediaMeta(ctx context.Context, id, name, altEN, altID string) error
	GetMedia(ctx context.Context, id string) (*MediaRecord, error)
	MediaBySlug(ctx context.Context, slug string) (*MediaRecord, error)
	ListMedia(ctx context.Context, q, kind string) ([]MediaRecord, error)
	MediaUsage(ctx context.Context, id string) (int, error)
	DeleteMedia(ctx context.Context, id string) error

	// Public site.
	PublicSnapshot(ctx context.Context) (*Snapshot, error)

	// Submissions.
	InsertQuotation(ctx context.Context, q QuotationRecord) error
	InsertApplication(ctx context.Context, a ApplicationRecord) error
	CountRecentSubmissions(ctx context.Context, table, column, value string, since time.Time) (int, error)
	ListQuotations(ctx context.Context, q ListQuery) ([]QuotationRecord, int, error)
	GetQuotation(ctx context.Context, id string) (*QuotationRecord, error)
	ListApplications(ctx context.Context, q ListQuery) ([]ApplicationRecord, int, error)
	GetApplication(ctx context.Context, id string) (*ApplicationRecord, error)
	ApplicationsOlderThan(ctx context.Context, cutoff time.Time) ([]ApplicationRecord, error)
	DeleteApplication(ctx context.Context, id string) error

	// Outbound mail.
	EnqueueMail(ctx context.Context, m OutboxMail) error
	DueMail(ctx context.Context, limit int, now time.Time) ([]OutboxMail, error)
	MarkMailSent(ctx context.Context, id string) error
	MarkMailFailed(ctx context.Context, id string, errText string, next time.Time) error
	ListOutbox(ctx context.Context, limit int) ([]OutboxMail, error)

	// Certificates.
	CertificatesForReminder(ctx context.Context) ([]CertificateRecord, error)
	RecordCertificateReminder(ctx context.Context, certID, validUntil string) (bool, error)
}

// Storage is the object store port (MinIO / S3).
type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, int64, error)
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration, downloadName string) (string, error)
}

// Mailer is the outbound mail port. SMTP in development and production; a
// second implementation (log only) runs when notifications are disabled.
type Mailer interface {
	Send(ctx context.Context, m Mail) error
}

type Mail struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// ImageProcessor decodes, checks and re-encodes an uploaded raster image.
type ImageProcessor interface {
	Process(b []byte, widths []int, maxDim int) (*ProcessedImage, error)
}

type ProcessedImage struct {
	Width, Height int
	Variants      []EncodedVariant
}

type EncodedVariant struct {
	Width int
	Bytes []byte
}

// ---------------------------------------------------------------------------
// Records
// ---------------------------------------------------------------------------

type User struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	Name              string     `json:"name"`
	PasswordHash      string     `json:"-"`
	Role              string     `json:"role"`
	Active            bool       `json:"active"`
	FailedLogins      int        `json:"-"`
	LockedUntil       *time.Time `json:"lockedUntil,omitempty"`
	LastLoginAt       *time.Time `json:"lastLoginAt,omitempty"`
	PasswordChangedAt time.Time  `json:"passwordChangedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type RefreshToken struct {
	ID         string
	UserID     string
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy string
	IP         string
	UserAgent  string
}

type AuditEntry struct {
	ID         string    `json:"id"`
	At         time.Time `json:"at"`
	ActorID    string    `json:"actorId,omitempty"`
	ActorEmail string    `json:"actorEmail"`
	Action     string    `json:"action"`
	Entity     string    `json:"entity"`
	EntityID   string    `json:"entityId"`
	Summary    string    `json:"summary"`
	Before     any       `json:"before,omitempty"`
	After      any       `json:"after,omitempty"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"userAgent"`
	RequestID  string    `json:"requestId"`
}

type Parameter struct {
	ID            string    `json:"id"`
	Group         string    `json:"grp"`
	Key           string    `json:"key"`
	ValueEN       string    `json:"value_en"`
	ValueID       string    `json:"value_id"`
	Description   string    `json:"description"`
	IsSecret      bool      `json:"is_secret"`
	IsPlaceholder bool      `json:"is_placeholder"`
	Sort          int       `json:"sort"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MediaVariant struct {
	W     int    `json:"w"`
	Key   string `json:"key"`
	Bytes int64  `json:"bytes"`
}

type MediaRecord struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug,omitempty"`
	Kind      string         `json:"kind"`
	Name      string         `json:"name"`
	AltEN     string         `json:"altEn"`
	AltID     string         `json:"altId"`
	Hash      string         `json:"hash,omitempty"`
	Mime      string         `json:"mime"`
	Width     int            `json:"width"`
	Height    int            `json:"height"`
	Bytes     int64          `json:"bytes"`
	ObjectKey string         `json:"-"`
	Variants  []MediaVariant `json:"variants"`
	CreatedAt time.Time      `json:"createdAt"`
	Usage     int            `json:"usage"`
}

type QuotationRecord struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	Name            string    `json:"name"`
	Company         string    `json:"company"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	BuildingTypeKey string    `json:"buildingTypeKey"`
	BuildingType    string    `json:"buildingType"`
	AreaM2          *int      `json:"areaM2"`
	Location        string    `json:"location"`
	Services        []string  `json:"services"`
	Message         string    `json:"message"`
	Lang            string    `json:"lang"`
	IP              string    `json:"-"`
	UserAgent       string    `json:"-"`
}

type ApplicationRecord struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Position   string    `json:"position"`
	OpeningID  string    `json:"openingId,omitempty"`
	Message    string    `json:"message"`
	CVKey      string    `json:"-"`
	CVFilename string    `json:"cvFilename"`
	CVMime     string    `json:"cvMime"`
	CVBytes    int64     `json:"cvBytes"`
	Lang       string    `json:"lang"`
	IP         string    `json:"-"`
	UserAgent  string    `json:"-"`
}

type OutboxMail struct {
	ID         string     `json:"id"`
	CreatedAt  time.Time  `json:"createdAt"`
	Kind       string     `json:"kind"`
	RefID      string     `json:"refId"`
	Recipients []string   `json:"recipients"`
	Subject    string     `json:"subject"`
	Body       string     `json:"-"`
	Attempts   int        `json:"attempts"`
	NextAt     time.Time  `json:"nextAttemptAt"`
	SentAt     *time.Time `json:"sentAt"`
	LastError  string     `json:"lastError"`
}

type CertificateRecord struct {
	ID         string
	Name       string
	ValidUntil string // YYYY-MM-DD or ""
	Published  bool
}

// Options feeds the CMS pickers: every row a Ref or Param field can point at.
type Options struct {
	Refs   map[string][]Option `json:"refs"`
	Params map[string][]Option `json:"params"`
	Groups []Option            `json:"groups"`
}

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
