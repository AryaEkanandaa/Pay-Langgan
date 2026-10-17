-- +goose Up
ALTER TABLE add_ons ADD COLUMN visible_in_checkout BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE add_ons DROP COLUMN visible_in_checkout;
