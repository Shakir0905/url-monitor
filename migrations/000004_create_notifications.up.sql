CREATE TABLE notifications (
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id    BIGINT NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    channel   VARCHAR(20) NOT NULL,
    status    VARCHAR(20) NOT NULL,
    payload   JSONB,
    sent_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_sent ON notifications(user_id, sent_at DESC);