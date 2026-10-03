package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pay-langgan/internal/database"
	"pay-langgan/internal/utils"
)

type BusinessGatewayInput struct {
	Provider            string `json:"provider"`
	Environment         string `json:"environment"`
	XenditSecretKey     string `json:"xendit_secret_key,omitempty"`
	XenditCallbackToken string `json:"xendit_callback_token,omitempty"`
	MidtransMerchantID  string `json:"midtrans_merchant_id,omitempty"`
	MidtransClientKey   string `json:"midtrans_client_key,omitempty"`
	MidtransServerKey   string `json:"midtrans_server_key,omitempty"`
}

type BusinessGatewaySummary struct {
	Provider     string     `db:"provider" json:"provider"`
	Environment  string     `db:"environment" json:"environment"`
	ConfiguredAt time.Time  `db:"configured_at" json:"configured_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	VerifiedAt   *time.Time `db:"verified_at" json:"verified_at"`
	AccountLabel string     `json:"account_label,omitempty"`
	WebhookURL   string     `json:"webhook_url"`
	Status       string     `json:"status"`
}

type BusinessGatewayService struct {
	db          *database.DB
	vault       *credentialVault
	webhookBase string
}

type BusinessGatewayRuntime struct {
	BusinessID string
	WebhookID  string
	Config     BusinessGatewayInput
}

func NewBusinessGatewayService(db *database.DB, encryptionKey, webhookBase string) *BusinessGatewayService {
	vault, _ := newCredentialVault(encryptionKey)
	return &BusinessGatewayService{db: db, vault: vault, webhookBase: strings.TrimRight(webhookBase, "/")}
}

func validateBusinessGateway(input BusinessGatewayInput) (BusinessGatewayInput, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.Environment = strings.ToLower(strings.TrimSpace(input.Environment))
	if input.Environment != "sandbox" && input.Environment != "production" {
		return input, fmt.Errorf("%w: choose sandbox or production mode", utils.ErrBadRequest)
	}
	switch input.Provider {
	case "xendit":
		input.XenditSecretKey = strings.TrimSpace(input.XenditSecretKey)
		input.XenditCallbackToken = strings.TrimSpace(input.XenditCallbackToken)
		if input.XenditSecretKey == "" || input.XenditCallbackToken == "" || len(input.XenditSecretKey) > 512 || len(input.XenditCallbackToken) > 512 {
			return input, fmt.Errorf("%w: Xendit secret key and callback token are required", utils.ErrBadRequest)
		}
		input.MidtransMerchantID, input.MidtransClientKey, input.MidtransServerKey = "", "", ""
	case "midtrans":
		input.MidtransMerchantID = strings.TrimSpace(input.MidtransMerchantID)
		input.MidtransClientKey = strings.TrimSpace(input.MidtransClientKey)
		input.MidtransServerKey = strings.TrimSpace(input.MidtransServerKey)
		if input.MidtransMerchantID == "" || input.MidtransClientKey == "" || input.MidtransServerKey == "" || len(input.MidtransMerchantID) > 128 || len(input.MidtransClientKey) > 512 || len(input.MidtransServerKey) > 512 {
			return input, fmt.Errorf("%w: Midtrans Merchant ID, Client Key, and Server Key are required", utils.ErrBadRequest)
		}
		input.XenditSecretKey, input.XenditCallbackToken = "", ""
	default:
		return input, fmt.Errorf("%w: choose Xendit or Midtrans", utils.ErrBadRequest)
	}
	return input, nil
}

func (s *BusinessGatewayService) Save(ctx context.Context, businessID string, input BusinessGatewayInput) (*BusinessGatewaySummary, error) {
	if s.db == nil || s.vault == nil {
		return nil, fmt.Errorf("%w: payment gateway credential encryption is not configured", utils.ErrConflict)
	}
	if businessID == "" {
		return nil, utils.ErrBadRequest
	}
	input, err := validateBusinessGateway(input)
	if err != nil {
		return nil, err
	}
	plain, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	ciphertext, err := s.vault.seal(plain)
	if err != nil {
		return nil, err
	}
	newWebhookID, err := secureToken(32)
	if err != nil {
		return nil, err
	}
	var saved BusinessGatewaySummary
	var webhookID string
	err = s.db.QueryRowxContext(ctx, `INSERT INTO business_payment_gateways
 (business_id,provider,environment,encrypted_credentials,webhook_id,configured_at,updated_at,verified_at)
 VALUES ($1,$2,$3,$4,$5,(NOW() AT TIME ZONE 'UTC'),(NOW() AT TIME ZONE 'UTC'),NULL)
 ON CONFLICT (business_id) DO UPDATE SET provider=EXCLUDED.provider,environment=EXCLUDED.environment,
 encrypted_credentials=EXCLUDED.encrypted_credentials,updated_at=(NOW() AT TIME ZONE 'UTC'),verified_at=NULL
 RETURNING provider,environment,configured_at,updated_at,verified_at,webhook_id`,
		businessID, input.Provider, input.Environment, ciphertext, newWebhookID).Scan(
		&saved.Provider, &saved.Environment, &saved.ConfiguredAt, &saved.UpdatedAt, &saved.VerifiedAt, &webhookID)
	if err != nil {
		return nil, fmt.Errorf("save business payment gateway: %w", err)
	}
	saved.WebhookURL = s.webhookBase + "/api/v1/webhooks/business-gateways/" + webhookID
	saved.Status = "configured"
	if saved.VerifiedAt != nil {
		saved.Status = "verified"
	}
	if input.Provider == "midtrans" {
		saved.AccountLabel = maskCredential(input.MidtransMerchantID)
	}
	return &saved, nil
}

func (s *BusinessGatewayService) Get(ctx context.Context, businessID string) (*BusinessGatewaySummary, error) {
	if s.db == nil || businessID == "" {
		return nil, utils.ErrBadRequest
	}
	var result BusinessGatewaySummary
	var webhookID string
	err := s.db.QueryRowxContext(ctx, `SELECT provider,environment,configured_at,updated_at,verified_at,webhook_id
 FROM business_payment_gateways WHERE business_id=$1`, businessID).
		Scan(&result.Provider, &result.Environment, &result.ConfiguredAt, &result.UpdatedAt, &result.VerifiedAt, &webhookID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get business payment gateway: %w", err)
	}
	result.WebhookURL = s.webhookBase + "/api/v1/webhooks/business-gateways/" + webhookID
	result.Status = "configured"
	if result.VerifiedAt != nil {
		result.Status = "verified"
	}
	if result.Provider == "midtrans" {
		var encrypted []byte
		if err := s.db.GetContext(ctx, &encrypted, `SELECT encrypted_credentials FROM business_payment_gateways WHERE business_id=$1`, businessID); err != nil {
			return nil, err
		}
		if s.vault != nil {
			plain, err := s.vault.open(encrypted)
			if err == nil {
				var secrets BusinessGatewayInput
				if json.Unmarshal(plain, &secrets) == nil {
					result.AccountLabel = maskCredential(secrets.MidtransMerchantID)
				}
			}
		}
	}
	return &result, nil
}

func (s *BusinessGatewayService) Delete(ctx context.Context, businessID string) error {
	if s.db == nil || businessID == "" {
		return utils.ErrBadRequest
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM business_payment_gateways WHERE business_id=$1`, businessID)
	if err != nil {
		return fmt.Errorf("remove business payment gateway: %w", err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return utils.ErrNotFound
	}
	return nil
}

func (s *BusinessGatewayService) LoadForBusiness(ctx context.Context, businessID string) (*BusinessGatewayRuntime, error) {
	return s.load(ctx, `business_id=$1`, businessID)
}

func (s *BusinessGatewayService) LoadForWebhook(ctx context.Context, webhookID string) (*BusinessGatewayRuntime, error) {
	return s.load(ctx, `webhook_id=$1`, webhookID)
}

func (s *BusinessGatewayService) load(ctx context.Context, where string, value string) (*BusinessGatewayRuntime, error) {
	if s.db == nil || s.vault == nil || value == "" {
		return nil, utils.ErrNotFound
	}
	var runtime BusinessGatewayRuntime
	var encrypted []byte
	query := `SELECT business_id,webhook_id,encrypted_credentials FROM business_payment_gateways WHERE ` + where
	if err := s.db.QueryRowxContext(ctx, query, value).Scan(&runtime.BusinessID, &runtime.WebhookID, &encrypted); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("load business payment gateway: %w", err)
	}
	plain, err := s.vault.open(encrypted)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(plain, &runtime.Config); err != nil {
		return nil, fmt.Errorf("decode business payment gateway credentials: %w", err)
	}
	return &runtime, nil
}

func maskCredential(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 4 {
		return "••••"
	}
	return strings.Repeat("•", len(value)-4) + value[len(value)-4:]
}
