CREATE TABLE urls (
    id                     BIGSERIAL PRIMARY KEY,
    user_id                BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url                    TEXT NOT NULL,
    check_interval_seconds INT NOT NULL DEFAULT 60,
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_checked_at        TIMESTAMPTZ
);

CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE INDEX idx_urls_active_check ON urls(is_active, last_checked_at) WHERE is_active = TRUE;
