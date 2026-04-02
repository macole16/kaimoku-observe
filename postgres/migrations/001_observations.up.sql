CREATE SEQUENCE IF NOT EXISTS obs_ref_seq START 1;
CREATE SEQUENCE IF NOT EXISTS faq_ref_seq START 1;

CREATE TABLE IF NOT EXISTS observations (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ref                  TEXT NOT NULL UNIQUE,
    user_id              BIGINT NOT NULL,
    domain_id            BIGINT NOT NULL,
    title                TEXT NOT NULL,
    description          TEXT NOT NULL,
    category_user        TEXT NOT NULL DEFAULT 'something_off',
    severity_user        TEXT,
    location             TEXT,
    auto_context         JSONB NOT NULL DEFAULT '{}',
    triage               JSONB,
    status               TEXT NOT NULL DEFAULT 'submitted',
    resolution_internal  TEXT,
    resolution_user      TEXT,
    related_observations JSONB,
    faq_article_id       BIGINT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_observations_user_id ON observations(user_id);
CREATE INDEX IF NOT EXISTS idx_observations_domain_id ON observations(domain_id);
CREATE INDEX IF NOT EXISTS idx_observations_status ON observations(status);
CREATE INDEX IF NOT EXISTS idx_observations_created_at ON observations(created_at DESC);

CREATE TABLE IF NOT EXISTS observation_screenshots (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    observation_id  BIGINT NOT NULL REFERENCES observations(id) ON DELETE CASCADE,
    filename        TEXT NOT NULL,
    content_type    TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL,
    storage_key     TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS observation_timeline (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    observation_id  BIGINT NOT NULL REFERENCES observations(id) ON DELETE CASCADE,
    event_type      TEXT NOT NULL,
    actor           TEXT NOT NULL,
    content         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_observation_timeline_obs ON observation_timeline(observation_id, created_at);

CREATE TABLE IF NOT EXISTS faq_articles (
    id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ref                 TEXT NOT NULL UNIQUE,
    title               TEXT NOT NULL,
    category            TEXT NOT NULL,
    tags                JSONB NOT NULL DEFAULT '[]',
    problem             TEXT NOT NULL,
    resolution          TEXT NOT NULL,
    related_articles    JSONB,
    observation_count   INT NOT NULL DEFAULT 0,
    deflection_count    INT NOT NULL DEFAULT 0,
    published_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_triggered_at   TIMESTAMPTZ,
    improvement_flagged BOOLEAN NOT NULL DEFAULT false,
    views               INT NOT NULL DEFAULT 0,
    helpful_yes         INT NOT NULL DEFAULT 0,
    helpful_no          INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_faq_articles_category ON faq_articles(category);

CREATE TABLE IF NOT EXISTS observation_notifications (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    observation_id  BIGINT REFERENCES observations(id) ON DELETE SET NULL,
    event_type      TEXT NOT NULL,
    title           TEXT NOT NULL,
    body            TEXT,
    is_read         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_obs_notifications_user ON observation_notifications(user_id, is_read, created_at DESC);
