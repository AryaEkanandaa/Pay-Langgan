-- Customer-authorized trial enrollment has no authenticated dashboard user.
-- Preserve UpdatedBy for business-managed changes while allowing consent
-- captured directly from the customer checkout.
-- +goose Up
ALTER TABLE subscription_autopay ALTER COLUMN updated_by DROP NOT NULL;

-- +goose Down
-- Existing customer-consented rows have no dashboard actor; rollback therefore
-- requires assigning them to a valid user before restoring NOT NULL.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM subscription_autopay WHERE updated_by IS NULL) THEN
        RAISE EXCEPTION 'Cannot restore subscription_autopay.updated_by NOT NULL while customer consent rows exist';
    END IF;
END $$;
ALTER TABLE subscription_autopay ALTER COLUMN updated_by SET NOT NULL;
