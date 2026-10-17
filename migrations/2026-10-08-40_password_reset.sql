-- +goose Up
ALTER TABLE users ADD COLUMN password_version INTEGER NOT NULL DEFAULT 0;
CREATE TABLE password_reset_requests (
 id BIGSERIAL PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 recipient VARCHAR(100) NOT NULL,
 token_hash BYTEA NOT NULL UNIQUE,
 nonce BYTEA,
 status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','sending','sent','failed','cancelled')),
 attempts INTEGER NOT NULL DEFAULT 0,
 next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW()+INTERVAL '15 minutes',
 used_at TIMESTAMPTZ
);
CREATE INDEX idx_password_reset_user_recent ON password_reset_requests(user_id,created_at DESC);
CREATE INDEX idx_password_reset_queue ON password_reset_requests(next_attempt_at,id) WHERE status IN ('queued','sending');

-- +goose Down
DROP TABLE password_reset_requests;
ALTER TABLE users DROP COLUMN password_version;
