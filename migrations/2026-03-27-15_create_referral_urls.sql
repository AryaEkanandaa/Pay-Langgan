-- +goose Up
CREATE TABLE referral_urls (
    id SERIAL PRIMARY KEY,
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL UNIQUE,
    target_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS referral_urls;
