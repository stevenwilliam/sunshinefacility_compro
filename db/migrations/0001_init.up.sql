-- 0001_init — the whole schema.
--
-- The database enforces the invariants, not just the application
-- (CLAUDE.md §6). Each rule cites the BR id in docs/02-business-rules.md that
-- it implements, so a reader can go from a constraint to the reason for it.

-- ---------------------------------------------------------------------------
-- Users, sessions, audit
-- ---------------------------------------------------------------------------

CREATE TABLE users (
    id                  uuid PRIMARY KEY,
    email               text NOT NULL,
    name                text NOT NULL,
    password_hash       text NOT NULL,
    -- BR-8.1: exactly one role per user, from a fixed list. A new role is a
    -- migration and a permissions-matrix change, never a free-text value.
    role                text NOT NULL CHECK (role IN ('superadmin', 'editor', 'sales', 'hr')),
    active              boolean NOT NULL DEFAULT true,
    failed_logins       integer NOT NULL DEFAULT 0 CHECK (failed_logins >= 0),
    locked_until        timestamptz,
    last_login_at       timestamptz,
    password_changed_at timestamptz NOT NULL DEFAULT now(),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CHECK (email = lower(email) AND email LIKE '%_@_%'),
    CHECK (password_hash LIKE '$argon2id$%')
);
CREATE UNIQUE INDEX users_email_uq ON users (email);

CREATE TABLE refresh_tokens (
    id          uuid PRIMARY KEY,
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash  text NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    replaced_by uuid,
    ip          text NOT NULL DEFAULT '',
    user_agent  text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX refresh_tokens_user_idx ON refresh_tokens (user_id);

-- Access tokens are short-lived; a logout puts the jti here until it would
-- have expired anyway.
CREATE TABLE revoked_jti (
    jti        text PRIMARY KEY,
    expires_at timestamptz NOT NULL
);

-- BR-10.1: append-only. No UPDATE, no DELETE — the trigger refuses both, so
-- not even a superadmin with SQL access through the app can rewrite history.
CREATE TABLE audit_log (
    id          uuid PRIMARY KEY,
    at          timestamptz NOT NULL DEFAULT now(),
    actor_id    uuid,
    actor_email text NOT NULL,
    action      text NOT NULL,
    entity      text NOT NULL,
    entity_id   text NOT NULL DEFAULT '',
    summary     text NOT NULL DEFAULT '',
    before      jsonb,
    after       jsonb,
    ip          text NOT NULL DEFAULT '',
    user_agent  text NOT NULL DEFAULT '',
    request_id  text NOT NULL DEFAULT ''
);
CREATE INDEX audit_log_at_idx ON audit_log (at DESC);
CREATE INDEX audit_log_entity_idx ON audit_log (entity, entity_id);

CREATE FUNCTION audit_log_append_only() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is append-only (BR-10.1)' USING ERRCODE = 'insufficient_privilege';
END $$;
CREATE TRIGGER audit_log_no_update BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_append_only();
CREATE TRIGGER audit_log_no_truncate BEFORE TRUNCATE ON audit_log
    FOR EACH STATEMENT EXECUTE FUNCTION audit_log_append_only();

-- ---------------------------------------------------------------------------
-- Media library
-- ---------------------------------------------------------------------------

CREATE TABLE media (
    id          uuid PRIMARY KEY,
    -- Stable handle for seeded files (e.g. 'photo/kitchen'); NULL for uploads.
    slug        text UNIQUE,
    kind        text NOT NULL CHECK (kind IN ('photo', 'logo', 'certificate', 'brand')),
    name        text NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    alt_en      text NOT NULL DEFAULT '',
    alt_id      text NOT NULL DEFAULT '',
    -- Filled when the bytes are processed. A row with no hash has no file yet
    -- and is never rendered (the public bundle skips it).
    hash        text UNIQUE,
    mime        text NOT NULL DEFAULT '' CHECK (mime IN ('', 'image/png', 'image/jpeg', 'image/webp', 'image/svg+xml')),
    width       integer NOT NULL DEFAULT 0 CHECK (width >= 0),
    height      integer NOT NULL DEFAULT 0 CHECK (height >= 0),
    bytes       bigint NOT NULL DEFAULT 0 CHECK (bytes >= 0),
    object_key  text NOT NULL DEFAULT '',
    -- [{"w":480,"key":"…","bytes":1234}, …] — the responsive WebP renditions.
    variants    jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_by  uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    -- BR-4.4: SVG only ever arrives through the seed (the brand marks). An
    -- upload is always a raster that was re-encoded.
    CHECK (mime <> 'image/svg+xml' OR slug IS NOT NULL)
);

-- ---------------------------------------------------------------------------
-- Configuration and copy
-- ---------------------------------------------------------------------------

-- CLAUDE.md §8: anything the business may change without a deploy.
CREATE TABLE parameters (
    id             uuid PRIMARY KEY,
    grp            text NOT NULL CHECK (grp ~ '^[a-z][a-z0-9_]*$'),
    key            text NOT NULL CHECK (key ~ '^[a-z0-9][a-z0-9_.-]*$'),
    value_en       text NOT NULL DEFAULT '',
    value_id       text NOT NULL DEFAULT '',
    description    text NOT NULL DEFAULT '',
    is_secret      boolean NOT NULL DEFAULT false,
    is_placeholder boolean NOT NULL DEFAULT false,
    sort           integer NOT NULL DEFAULT 0,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL,
    UNIQUE (grp, key)
);

-- Every interface string of the public site (BR-1.1): nav labels, buttons,
-- kickers, form labels, validation messages. One row, both languages.
CREATE TABLE site_text (
    key            text PRIMARY KEY CHECK (key ~ '^[a-z][a-zA-Z0-9_.]*$'),
    value_en       text NOT NULL,
    value_id       text NOT NULL,
    note           text NOT NULL DEFAULT '',
    is_placeholder boolean NOT NULL DEFAULT false,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL
);

-- The README's "page section": a titled block of a page, with body and image.
CREATE TABLE page_sections (
    id             uuid PRIMARY KEY,
    page           text NOT NULL CHECK (page IN ('home', 'services', 'service_detail', 'industries', 'about', 'careers', 'contact')),
    section        text NOT NULL CHECK (section ~ '^[a-z][a-z0-9_]*$'),
    title_en       text NOT NULL DEFAULT '',
    title_id       text NOT NULL DEFAULT '',
    body_en        text NOT NULL DEFAULT '',
    body_id        text NOT NULL DEFAULT '',
    image_id       uuid REFERENCES media (id) ON DELETE RESTRICT,
    published      boolean NOT NULL DEFAULT true,
    is_placeholder boolean NOT NULL DEFAULT false,
    sort           integer NOT NULL DEFAULT 0,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL,
    UNIQUE (page, section)
);

-- ---------------------------------------------------------------------------
-- Business lines
-- ---------------------------------------------------------------------------

CREATE TABLE service_lines (
    id             uuid PRIMARY KEY,
    key            text NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9-]{1,30}$'),
    sort           integer NOT NULL DEFAULT 0,
    title_en       text NOT NULL,
    title_id       text NOT NULL,
    -- Short chip label ("Food", "Digital") and the filter label
    -- ("Digital & Operational") used on Careers, Contact and Industries.
    short_en       text NOT NULL DEFAULT '',
    short_id       text NOT NULL DEFAULT '',
    label_en       text NOT NULL DEFAULT '',
    label_id       text NOT NULL DEFAULT '',
    summary_en     text NOT NULL DEFAULT '',  -- Home card and hero detail panel
    summary_id     text NOT NULL DEFAULT '',
    description_en text NOT NULL DEFAULT '',  -- Services page row
    description_id text NOT NULL DEFAULT '',
    lead_en        text NOT NULL DEFAULT '',  -- Service Detail hero
    lead_id        text NOT NULL DEFAULT '',
    color_base     text NOT NULL CHECK (color_base ~ '^#[0-9A-F]{6}$'),
    color_ink      text NOT NULL CHECK (color_ink ~ '^#[0-9A-F]{6}$'),
    color_shade    text NOT NULL CHECK (color_shade ~ '^#[0-9A-F]{6}$'),
    color_on_ink   text NOT NULL CHECK (color_on_ink ~ '^#[0-9A-F]{6}$'),
    color_label    text NOT NULL CHECK (color_label ~ '^#[0-9A-F]{6}$'),
    mark_id        uuid REFERENCES media (id) ON DELETE RESTRICT,
    photo_id       uuid REFERENCES media (id) ON DELETE RESTRICT,
    show_map       boolean NOT NULL DEFAULT false,
    show_certs     boolean NOT NULL DEFAULT false,
    map_image_id   uuid REFERENCES media (id) ON DELETE RESTRICT,
    published      boolean NOT NULL DEFAULT true,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL
);

-- The short lists on each line: the 4 tags on a Home card, the 6 tags on the
-- Services row, and the 4 "What's included" points in the hero panel.
CREATE TABLE service_items (
    id         uuid PRIMARY KEY,
    line_id    uuid NOT NULL REFERENCES service_lines (id) ON DELETE CASCADE,
    kind       text NOT NULL CHECK (kind IN ('card_tag', 'service_tag', 'hero_point')),
    text_en    text NOT NULL,
    text_id    text NOT NULL,
    sort       integer NOT NULL DEFAULT 0,
    published  boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES users (id) ON DELETE SET NULL
);
CREATE INDEX service_items_line_idx ON service_items (line_id, kind, sort);

CREATE TABLE service_facts (
    id         uuid PRIMARY KEY,
    line_id    uuid NOT NULL REFERENCES service_lines (id) ON DELETE CASCADE,
    value_en   text NOT NULL,
    value_id   text NOT NULL,
    label_en   text NOT NULL,
    label_id   text NOT NULL,
    sort       integer NOT NULL DEFAULT 0,
    published  boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES users (id) ON DELETE SET NULL
);
CREATE INDEX service_facts_line_idx ON service_facts (line_id, sort);

-- "What's included" cards on Service Detail: a list of sub-items OR a
-- paragraph. Sub-items are one per line in items_en / items_id.
CREATE TABLE service_groups (
    id         uuid PRIMARY KEY,
    line_id    uuid NOT NULL REFERENCES service_lines (id) ON DELETE CASCADE,
    title_en   text NOT NULL,
    title_id   text NOT NULL,
    body_en    text NOT NULL DEFAULT '',
    body_id    text NOT NULL DEFAULT '',
    items_en   text NOT NULL DEFAULT '',
    items_id   text NOT NULL DEFAULT '',
    sort       integer NOT NULL DEFAULT 0,
    published  boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES users (id) ON DELETE SET NULL
);
CREATE INDEX service_groups_line_idx ON service_groups (line_id, sort);

-- ---------------------------------------------------------------------------
-- Industries and clients
-- ---------------------------------------------------------------------------

CREATE TABLE industries (
    id          uuid PRIMARY KEY,
    key         text NOT NULL UNIQUE CHECK (key ~ '^[a-z][a-z0-9-]{1,30}$'),
    icon        text NOT NULL DEFAULT '' CHECK (icon ~ '^[a-z0-9-]*$'),
    title_en    text NOT NULL,
    title_id    text NOT NULL,
    body_en     text NOT NULL DEFAULT '',
    body_id     text NOT NULL DEFAULT '',
    sort        integer NOT NULL DEFAULT 0,
    published   boolean NOT NULL DEFAULT true,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    updated_by  uuid REFERENCES users (id) ON DELETE SET NULL
);

-- One mapping of which lines serve which industry. It drives BOTH the
-- Industries page chips and the Service Detail "industries we serve" chips,
-- so the two pages can never disagree (decision D12).
CREATE TABLE industry_lines (
    industry_id uuid NOT NULL REFERENCES industries (id) ON DELETE CASCADE,
    line_id     uuid NOT NULL REFERENCES service_lines (id) ON DELETE CASCADE,
    sort        integer NOT NULL DEFAULT 0,
    PRIMARY KEY (industry_id, line_id)
);

CREATE TABLE clients (
    id          uuid PRIMARY KEY,
    name        text NOT NULL,
    logo_id     uuid REFERENCES media (id) ON DELETE RESTRICT,
    industry_id uuid REFERENCES industries (id) ON DELETE SET NULL,
    sort        integer NOT NULL DEFAULT 0,
    published   boolean NOT NULL DEFAULT true,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    updated_by  uuid REFERENCES users (id) ON DELETE SET NULL
);

-- ---------------------------------------------------------------------------
-- Certificates
-- ---------------------------------------------------------------------------

CREATE TABLE certificates (
    id             uuid PRIMARY KEY,
    name           text NOT NULL,
    issuer         text NOT NULL DEFAULT '',
    valid_until    date,
    image_id       uuid REFERENCES media (id) ON DELETE RESTRICT,
    description_en text NOT NULL DEFAULT '',
    description_id text NOT NULL DEFAULT '',
    sort           integer NOT NULL DEFAULT 0,
    published      boolean NOT NULL DEFAULT false,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL
);

-- BR-3.2: an expired certificate cannot be published. "Today" is the
-- business day in Asia/Jakarta, stated explicitly rather than inherited from
-- the server's clock zone (CLAUDE.md §6). A CHECK cannot use the date (it is
-- not immutable), so a trigger refuses the write — including a direct UPDATE
-- that bypasses the API.
CREATE FUNCTION certificates_refuse_expired_publish() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.published AND NEW.valid_until IS NOT NULL
       AND NEW.valid_until < (now() AT TIME ZONE 'Asia/Jakarta')::date THEN
        RAISE EXCEPTION 'certificate % expired on % and cannot be published (BR-3.2)', NEW.name, NEW.valid_until
            USING ERRCODE = 'check_violation', CONSTRAINT = 'certificates_expired_unpublished';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER certificates_expired_guard BEFORE INSERT OR UPDATE ON certificates
    FOR EACH ROW EXECUTE FUNCTION certificates_refuse_expired_publish();

-- BR-3.4: one reminder per certificate per expiry date. Renewing the
-- certificate (a new valid_until) makes it eligible again.
CREATE TABLE certificate_reminders (
    certificate_id uuid NOT NULL REFERENCES certificates (id) ON DELETE CASCADE,
    valid_until    date NOT NULL,
    sent_at        timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (certificate_id, valid_until)
);

-- ---------------------------------------------------------------------------
-- Careers, testimonials, people, lists, stats
-- ---------------------------------------------------------------------------

CREATE TABLE career_openings (
    id              uuid PRIMARY KEY,
    title_en        text NOT NULL,
    title_id        text NOT NULL,
    line_id         uuid NOT NULL REFERENCES service_lines (id) ON DELETE RESTRICT,
    -- Location and type are parameter keys. The constant group column lets a
    -- composite FK prove the key exists in the right group.
    location_grp    text NOT NULL GENERATED ALWAYS AS ('locations') STORED,
    location_key    text NOT NULL,
    type_grp        text NOT NULL GENERATED ALWAYS AS ('job_types') STORED,
    type_key        text NOT NULL,
    closing_date    date,
    summary_en      text NOT NULL DEFAULT '',
    summary_id      text NOT NULL DEFAULT '',
    requirements_en text NOT NULL DEFAULT '',  -- one per line
    requirements_id text NOT NULL DEFAULT '',
    sort            integer NOT NULL DEFAULT 0,
    published       boolean NOT NULL DEFAULT true,
    is_placeholder  boolean NOT NULL DEFAULT false,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    updated_by      uuid REFERENCES users (id) ON DELETE SET NULL,
    FOREIGN KEY (location_grp, location_key) REFERENCES parameters (grp, key) ON DELETE RESTRICT,
    FOREIGN KEY (type_grp, type_key) REFERENCES parameters (grp, key) ON DELETE RESTRICT
);

CREATE TABLE testimonials (
    id             uuid PRIMARY KEY,
    quote_en       text NOT NULL,
    quote_id       text NOT NULL,
    name           text NOT NULL,
    role           text NOT NULL DEFAULT '',
    company        text NOT NULL,
    sort           integer NOT NULL DEFAULT 0,
    published      boolean NOT NULL DEFAULT false,
    is_placeholder boolean NOT NULL DEFAULT false,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL
);

CREATE TABLE management (
    id             uuid PRIMARY KEY,
    name           text NOT NULL,
    role_en        text NOT NULL,
    role_id        text NOT NULL,
    photo_id       uuid REFERENCES media (id) ON DELETE RESTRICT,
    sort           integer NOT NULL DEFAULT 0,
    published      boolean NOT NULL DEFAULT true,
    is_placeholder boolean NOT NULL DEFAULT false,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL
);

-- The small repeated lists that share one shape — icon, title, body, a meta
-- line (a year) and an accent colour. One table with an allow-listed list key
-- rather than eight identical tables (decision D11). Each list still gets its
-- own CMS screen.
CREATE TABLE content_items (
    id             uuid PRIMARY KEY,
    list_key       text NOT NULL CHECK (list_key IN (
                       'home_why', 'services_benefits', 'service_proof', 'careers_reasons',
                       'about_journey', 'about_mission', 'about_values', 'about_ecosystem')),
    icon           text NOT NULL DEFAULT '' CHECK (icon ~ '^[a-z0-9-]*$'),
    title_en       text NOT NULL DEFAULT '',
    title_id       text NOT NULL DEFAULT '',
    body_en        text NOT NULL DEFAULT '',
    body_id        text NOT NULL DEFAULT '',
    meta_en        text NOT NULL DEFAULT '',
    meta_id        text NOT NULL DEFAULT '',
    color          text NOT NULL DEFAULT '' CHECK (color = '' OR color ~ '^#[0-9A-F]{6}$'),
    sort           integer NOT NULL DEFAULT 0,
    published      boolean NOT NULL DEFAULT true,
    is_placeholder boolean NOT NULL DEFAULT false,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by     uuid REFERENCES users (id) ON DELETE SET NULL
);
CREATE INDEX content_items_list_idx ON content_items (list_key, sort);

CREATE TABLE stats (
    id         uuid PRIMARY KEY,
    scope      text NOT NULL CHECK (scope IN ('home', 'about')),
    value_en   text NOT NULL,
    value_id   text NOT NULL,
    label_en   text NOT NULL,
    label_id   text NOT NULL,
    sort       integer NOT NULL DEFAULT 0,
    published  boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES users (id) ON DELETE SET NULL
);

-- ---------------------------------------------------------------------------
-- Submissions and outbound mail
-- ---------------------------------------------------------------------------

CREATE TABLE quotation_requests (
    id                uuid PRIMARY KEY,
    created_at        timestamptz NOT NULL DEFAULT now(),
    name              text NOT NULL,
    company           text NOT NULL,
    email             text NOT NULL,
    phone             text NOT NULL,
    building_type_key text NOT NULL,
    building_type     text NOT NULL,          -- the label as the visitor saw it
    area_m2           integer CHECK (area_m2 IS NULL OR area_m2 BETWEEN 1 AND 100000000),
    location          text NOT NULL,
    services          text[] NOT NULL CHECK (cardinality(services) >= 1),
    message           text NOT NULL DEFAULT '',
    lang              text NOT NULL CHECK (lang IN ('en', 'id')),
    ip                text NOT NULL DEFAULT '',
    user_agent        text NOT NULL DEFAULT ''
);
CREATE INDEX quotation_requests_created_idx ON quotation_requests (created_at DESC);

CREATE TABLE career_applications (
    id           uuid PRIMARY KEY,
    created_at   timestamptz NOT NULL DEFAULT now(),
    name         text NOT NULL,
    phone        text NOT NULL,
    email        text NOT NULL,
    position     text NOT NULL,
    opening_id   uuid REFERENCES career_openings (id) ON DELETE SET NULL,
    message      text NOT NULL DEFAULT '',
    cv_key       text NOT NULL,
    cv_filename  text NOT NULL,
    cv_mime      text NOT NULL CHECK (cv_mime IN (
                     'application/pdf', 'application/msword',
                     'application/vnd.openxmlformats-officedocument.wordprocessingml.document')),
    cv_bytes     bigint NOT NULL CHECK (cv_bytes > 0),
    lang         text NOT NULL CHECK (lang IN ('en', 'id')),
    ip           text NOT NULL DEFAULT '',
    user_agent   text NOT NULL DEFAULT ''
);
CREATE INDEX career_applications_created_idx ON career_applications (created_at DESC);

-- Mail is written here in the same transaction as the thing it is about, and
-- a worker sends it. A submission is therefore never lost because the SMTP
-- relay was down for a minute (BR-5.3).
CREATE TABLE notification_outbox (
    id              uuid PRIMARY KEY,
    created_at      timestamptz NOT NULL DEFAULT now(),
    kind            text NOT NULL CHECK (kind IN ('quotation', 'application', 'certificate_expiry')),
    ref_id          text NOT NULL DEFAULT '',
    recipients      text[] NOT NULL CHECK (cardinality(recipients) >= 1),
    subject         text NOT NULL,
    body            text NOT NULL,
    attempts        integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    sent_at         timestamptz,
    last_error      text NOT NULL DEFAULT ''
);
CREATE INDEX notification_outbox_due_idx ON notification_outbox (next_attempt_at) WHERE sent_at IS NULL;
