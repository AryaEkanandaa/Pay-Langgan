-- +goose Up
ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS chk_checkout_sessions_provider;
ALTER TABLE checkout_sessions ADD CONSTRAINT chk_checkout_sessions_provider
    CHECK (provider IN ('xendit', 'midtrans'));

-- +goose Down
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM checkout_sessions WHERE provider = 'midtrans') THEN
        RAISE EXCEPTION 'Cannot remove Midtrans checkout support while Midtrans checkout sessions exist';
    END IF;
END $$;
ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS chk_checkout_sessions_provider;
ALTER TABLE checkout_sessions ADD CONSTRAINT chk_checkout_sessions_provider CHECK (provider IN ('xendit'));
