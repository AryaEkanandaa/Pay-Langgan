-- +goose Up
ALTER TABLE coupons ADD COLUMN business_id VARCHAR(50);

-- Existing coupons can only be assigned automatically when their subscription
-- usage points to exactly one business. Ambiguous legacy rows must be mapped
-- manually before this migration is applied.
DO $$
BEGIN
    IF (SELECT COUNT(*) FROM businesses) = 1 THEN
        UPDATE coupons
        SET business_id = (SELECT id FROM businesses LIMIT 1)
        WHERE business_id IS NULL;
    END IF;

    UPDATE coupons c
    SET business_id = scoped.business_id
    FROM (
        SELECT sc.coupon_id, MIN(cu.business_id) AS business_id
        FROM subscription_coupons sc
        JOIN subscriptions su ON su.id = sc.subscription_id
        JOIN customers cu ON cu.id = su.customer_id
        GROUP BY sc.coupon_id
        HAVING COUNT(DISTINCT cu.business_id) = 1
    ) scoped
    WHERE c.id = scoped.coupon_id AND c.business_id IS NULL;

    IF EXISTS (SELECT 1 FROM coupons WHERE business_id IS NULL) THEN
        RAISE EXCEPTION 'cannot scope existing coupons: assign business_id for every legacy coupon first';
    END IF;
END $$;

ALTER TABLE coupons
    ALTER COLUMN business_id SET NOT NULL,
    ADD CONSTRAINT fk_coupons_business FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;

ALTER TABLE coupons DROP CONSTRAINT IF EXISTS coupons_code_key;
CREATE UNIQUE INDEX ux_coupons_business_code ON coupons(business_id, code);
CREATE INDEX idx_coupons_business_id ON coupons(business_id);

-- +goose Down
DROP INDEX IF EXISTS idx_coupons_business_id;
DROP INDEX IF EXISTS ux_coupons_business_code;
ALTER TABLE coupons DROP CONSTRAINT IF EXISTS fk_coupons_business;
ALTER TABLE coupons ADD CONSTRAINT coupons_code_key UNIQUE (code);
ALTER TABLE coupons DROP COLUMN IF EXISTS business_id;
