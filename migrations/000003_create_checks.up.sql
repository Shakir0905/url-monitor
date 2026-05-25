CREATE TABLE checks (
    id               BIGSERIAL PRIMARY KEY,
    url_id           BIGINT NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    status_code      INT,
    response_time_ms INT,
    is_up            BOOLEAN NOT NULL,
    error_message    TEXT,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_checks_url_checked ON checks(url_id, checked_at DESC);
CREATE INDEX idx_checks_checked_at ON checks(checked_at);