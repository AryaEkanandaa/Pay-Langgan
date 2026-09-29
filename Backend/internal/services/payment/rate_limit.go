package payment

import (
	"context"
	"database/sql"
)

// Shared by all API instances. Only hashed identifiers are stored.
func (s *CheckoutService) AllowCheckoutRequest(ctx context.Context, key string, limit int) (bool, error) {
	var requests int
	err := s.db.QueryRowContext(ctx, `INSERT INTO checkout_rate_limits(bucket_key,window_start,requests)
 VALUES($1,(NOW() AT TIME ZONE 'UTC'),1)
 ON CONFLICT(bucket_key) DO UPDATE SET
 requests=CASE WHEN checkout_rate_limits.window_start <= (NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute' THEN 1 ELSE checkout_rate_limits.requests+1 END,
 window_start=CASE WHEN checkout_rate_limits.window_start <= (NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute' THEN (NOW() AT TIME ZONE 'UTC') ELSE checkout_rate_limits.window_start END
 WHERE checkout_rate_limits.window_start <= (NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute' OR checkout_rate_limits.requests < $2
 RETURNING requests`, hashToken(key), limit).Scan(&requests)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
