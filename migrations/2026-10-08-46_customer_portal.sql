-- +goose Up
CREATE TABLE customer_portal_requests (
 id BIGSERIAL PRIMARY KEY,
 site_id BIGINT NOT NULL REFERENCES merchant_embed_sites(id) ON DELETE CASCADE,
 recipient TEXT NOT NULL,
 token_hash BYTEA NOT NULL UNIQUE,
 nonce BYTEA,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW()+INTERVAL '15 minutes',
 used_at TIMESTAMPTZ,
 status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','sending','sent','failed','cancelled','previewed')),
 attempts INT NOT NULL DEFAULT 0,
 next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_portal_request_recipient ON customer_portal_requests(site_id,recipient,created_at);
CREATE INDEX idx_portal_request_queue ON customer_portal_requests(next_attempt_at) WHERE status IN ('queued','sending');
CREATE TABLE customer_portal_sessions (
 id BIGSERIAL PRIMARY KEY,
 site_id BIGINT NOT NULL REFERENCES merchant_embed_sites(id) ON DELETE CASCADE,
 recipient TEXT NOT NULL,
 token_hash BYTEA NOT NULL UNIQUE,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW()+INTERVAL '12 hours'
);
CREATE TABLE portal_checkout_access (
 token_hash VARCHAR(64) PRIMARY KEY,
 checkout_id BIGINT NOT NULL REFERENCES checkout_sessions(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_portal_checkout_expiry ON portal_checkout_access(expires_at);
-- +goose Down
DROP TABLE portal_checkout_access;
DROP TABLE customer_portal_sessions;
DROP TABLE customer_portal_requests;
