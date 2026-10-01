-- +goose Up
CREATE TABLE merchant_embed_sites (
    id BIGSERIAL PRIMARY KEY,
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    public_id VARCHAR(64) NOT NULL UNIQUE,
    origin VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    revoked_at TIMESTAMP
);
CREATE INDEX idx_merchant_embed_sites_business ON merchant_embed_sites(business_id, id DESC);

CREATE TABLE merchant_embed_site_plans (
    site_id BIGINT NOT NULL REFERENCES merchant_embed_sites(id) ON DELETE CASCADE,
    plan_id INT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    PRIMARY KEY (site_id, plan_id)
);

-- +goose Down
DROP TABLE merchant_embed_site_plans;
DROP TABLE merchant_embed_sites;
