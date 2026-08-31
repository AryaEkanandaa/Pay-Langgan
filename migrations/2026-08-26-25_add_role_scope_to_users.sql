-- +goose Up
ALTER TABLE users
    ALTER COLUMN business_id DROP NOT NULL;

ALTER TABLE users
    ADD CONSTRAINT chk_users_role
    CHECK (role IN ('super_admin', 'admin', 'sales', 'finance'));

ALTER TABLE users
    ADD CONSTRAINT chk_users_business_scope
    CHECK (
        (role = 'super_admin' AND business_id IS NULL)
        OR (role <> 'super_admin' AND business_id IS NOT NULL)
    );

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_business_scope;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users ALTER COLUMN business_id SET NOT NULL;
