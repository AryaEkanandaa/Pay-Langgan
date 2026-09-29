package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pay-langgan/internal/utils"
)

type AutopaySettings struct {
	SubscriptionID   int       `db:"subscription_id" json:"subscription_id"`
	PaymentMethodID  int       `db:"payment_method_id" json:"payment_method_id"`
	Enabled          bool      `db:"enabled" json:"enabled"`
	ConsentReference string    `db:"consent_reference" json:"consent_reference"`
	ConsentedAt      time.Time `db:"consented_at" json:"consented_at"`
	ConsentExpiresAt time.Time `db:"consent_expires_at" json:"consent_expires_at"`
	MaxAmount        int64     `db:"max_amount" json:"max_amount"`
	EffectiveAt      time.Time `db:"effective_at" json:"effective_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
	UpdatedBy        int       `db:"updated_by" json:"updated_by"`
}

type AutopayRequest struct {
	Enabled          bool      `json:"enabled"`
	PaymentMethodID  int       `json:"payment_method_id"`
	ConsentReference string    `json:"consent_reference"`
	ConsentedAt      time.Time `json:"consented_at"`
	ConsentExpiresAt time.Time `json:"consent_expires_at"`
	MaxAmount        int64     `json:"max_amount"`
}

func (s *CheckoutService) AutopaySettings(ctx context.Context, id int, businessID string) (*AutopaySettings, error) {
	var exists bool
	if err := s.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM subscriptions s JOIN customers c ON c.id=s.customer_id WHERE s.id=$1 AND c.business_id=$2)`, id, businessID); err != nil {
		return nil, err
	}
	if !exists {
		return nil, utils.ErrNotFound
	}
	var settings AutopaySettings
	err := s.db.GetContext(ctx, &settings, `SELECT * FROM subscription_autopay WHERE subscription_id=$1`, id)
	if err == sql.ErrNoRows {
		return &AutopaySettings{SubscriptionID: id}, nil
	}
	return &settings, err
}

func (s *CheckoutService) SetAutopay(ctx context.Context, id int, businessID string, userID int, r AutopayRequest) (*AutopaySettings, error) {
	if id < 1 || userID < 1 || businessID == "" {
		return nil, utils.ErrBadRequest
	}
	r.ConsentReference = strings.TrimSpace(r.ConsentReference)
	if r.Enabled && (r.PaymentMethodID < 1 || r.ConsentReference == "" || len(r.ConsentReference) > 255 ||
		r.ConsentedAt.IsZero() || r.ConsentedAt.After(time.Now()) || !r.ConsentExpiresAt.After(time.Now()) ||
		r.MaxAmount < 5000 || r.MaxAmount > 200000000) {
		return nil, fmt.Errorf("%w: payment method, customer consent evidence, valid dates and maximum IDR amount are required", utils.ErrBadRequest)
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var sub struct {
		CustomerID int     `db:"customer_id"`
		Status     string  `db:"status"`
		Cycle      *string `db:"billing_cycle"`
	}
	err = tx.GetContext(ctx, &sub, `SELECT s.customer_id,s.status,s.billing_cycle FROM subscriptions s JOIN customers c ON c.id=s.customer_id
 WHERE s.id=$1 AND c.business_id=$2 FOR UPDATE OF s`, id, businessID)
	if err == sql.ErrNoRows {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if r.Enabled {
		if sub.Status != "active" || sub.Cycle == nil || (*sub.Cycle != "monthly" && *sub.Cycle != "yearly") {
			return nil, fmt.Errorf("%w: autopay requires an active subscription with a billing schedule", utils.ErrConflict)
		}
		var valid bool
		if err := tx.GetContext(ctx, &valid, `SELECT EXISTS(SELECT 1 FROM payment_methods WHERE id=$1 AND customer_id=$2 AND provider='xendit')`, r.PaymentMethodID, sub.CustomerID); err != nil {
			return nil, err
		}
		if !valid {
			return nil, fmt.Errorf("%w: payment method does not belong to this customer", utils.ErrBadRequest)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO subscription_autopay(subscription_id,payment_method_id,enabled,consent_reference,consented_at,consent_expires_at,max_amount,updated_by)
 VALUES($1,$2,TRUE,$3,$4,$5,$6,$7) ON CONFLICT(subscription_id) DO UPDATE SET payment_method_id=$2,enabled=TRUE,
 consent_reference=$3,consented_at=$4,consent_expires_at=$5,max_amount=$6,updated_by=$7,
 effective_at=(NOW() AT TIME ZONE 'UTC'),updated_at=(NOW() AT TIME ZONE 'UTC')`, id, r.PaymentMethodID, r.ConsentReference, r.ConsentedAt.UTC(), r.ConsentExpiresAt.UTC(), r.MaxAmount, userID)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE subscription_autopay SET enabled=FALSE,updated_by=$2,updated_at=(NOW() AT TIME ZONE 'UTC') WHERE subscription_id=$1`, id, userID)
	}
	if err != nil {
		return nil, err
	}
	// Preserve each consent grant/revocation independently of the current setting.
	audit, _ := json.Marshal(r)
	_, err = tx.ExecContext(ctx, `INSERT INTO audit_logs(business_id,user_id,action,entity_type,entity_id,new_value)
 VALUES($1,$2,'set_autopay','subscription',$3,$4::jsonb)`, businessID, userID, fmt.Sprint(id), string(audit))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.AutopaySettings(ctx, id, businessID)
}
