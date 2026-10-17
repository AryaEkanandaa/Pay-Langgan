package payment

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type MidtransCardProvider interface {
	ChargeCard(context.Context, provider.MidtransCardRequest) (*provider.MidtransCardTransaction, error)
	GetCardTransaction(context.Context, string) (*provider.MidtransCardTransaction, error)
}

func (s *CheckoutService) SetMidtransCardProviderFactory(factory func(*BusinessGatewayRuntime) MidtransCardProvider) {
	s.midtransCardFactory = factory
}

// VerifyMidtransCard fetches the outcome from the merchant-authenticated API.
// A browser callback only requests verification; it never supplies payment status.
func (s *CheckoutService) VerifyMidtransCard(ctx context.Context, token string) (*models.PublicCheckoutResponse, error) {
	record, err := s.findByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if record.Provider == "midtrans" && record.CheckoutMode == "CORE" &&
		(record.Status == "active" || record.Status == "creating") && record.ProviderSessionID != nil {
		if err := s.reconcileRecord(ctx, record, *record.ProviderSessionID); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, token)
}

func (s *CheckoutService) midtransCardProvider(gateway *BusinessGatewayRuntime) MidtransCardProvider {
	if s.midtransCardFactory != nil {
		return s.midtransCardFactory(gateway)
	}
	origin := "https://api.sandbox.midtrans.com"
	if gateway.Config.Environment == "production" {
		origin = "https://api.midtrans.com"
	}
	return provider.NewMidtransProvider(gateway.Config.MidtransServerKey, gateway.Config.MidtransClientKey, origin, false)
}

func coreCheckoutResponse(record *models.CheckoutRecord, gateway *BusinessGatewayRuntime) *models.StartCheckoutResponse {
	result := &models.StartCheckoutResponse{Provider: "midtrans", Mode: "CORE", ClientKey: gateway.Config.MidtransClientKey,
		Environment: gateway.Config.Environment, ExpiresAt: record.ExpiresAt, Status: "active"}
	if record.ProviderSessionID != nil {
		result.PaymentSessionID = *record.ProviderSessionID
	}
	if record.AuthenticationURL != nil {
		result.AuthenticationURL = *record.AuthenticationURL
	}
	return result
}

func (s *CheckoutService) ChargeMidtransCard(ctx context.Context, token string, request models.MidtransCardCheckoutRequest) (*models.MidtransCardCheckoutResponse, error) {
	if strings.TrimSpace(request.TokenID) == "" || len(request.TokenID) > 255 {
		return nil, fmt.Errorf("%w: card token is required", utils.ErrBadRequest)
	}
	if err := validateCheckoutContact(request.Name, request.Email, request.Mobile); err != nil {
		return nil, err
	}
	record, err := s.findByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if record.Provider != "midtrans" || record.CheckoutMode != "CORE" || record.InvoiceStatus != "pending" || !record.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("%w: card checkout is unavailable", utils.ErrConflict)
	}
	gateway, err := s.gatewayForBusiness(ctx, record.BusinessID)
	if err != nil {
		return nil, err
	}
	if gateway == nil || gateway.Config.Provider != "midtrans" {
		return nil, fmt.Errorf("%w: payment gateway changed; create a new checkout", utils.ErrConflict)
	}
	if record.Status != "created" {
		return nil, fmt.Errorf("%w: payment has already started; wait for verification", utils.ErrConflict)
	}
	claimed, err := s.repo.Claim(ctx, record.ID, request.Email, request.Mobile)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, fmt.Errorf("%w: payment has already started", utils.ErrConflict)
	}
	// Core API supports looking up the immutable order ID after a lost response.
	// Persist it before making a charge so recovery never submits another charge.
	if err = s.repo.RememberSession(ctx, record.ID, record.ReferenceID); err != nil {
		return nil, err
	}
	transaction, chargeErr := s.midtransCardProvider(gateway).ChargeCard(ctx, provider.MidtransCardRequest{
		OrderID: record.ReferenceID, Amount: int64(record.InvoiceAmount), TokenID: request.TokenID, Name: request.Name,
		Email: request.Email, Phone: strings.TrimPrefix(request.Mobile, "+"),
	})
	recoveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if chargeErr != nil {
		var apiErr *provider.MidtransAPIError
		if errors.As(chargeErr, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 406 && apiErr.StatusCode != 409 {
			if err = s.repo.MarkFailed(recoveryCtx, record.ID); err != nil {
				return nil, err
			}
		} else if err = s.noteRecovery(recoveryCtx, record.ID, "card_charge_uncertain", true, time.Minute); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Midtrans card charge failed: %w", chargeErr)
	}
	if err = s.acceptMidtransCardTransaction(recoveryCtx, record, gateway, transaction); err != nil {
		_ = s.noteRecovery(recoveryCtx, record.ID, "card_response_unverified", true, time.Minute)
		return nil, err
	}
	return &models.MidtransCardCheckoutResponse{AuthenticationURL: transaction.RedirectURL, Status: transaction.TransactionStatus}, nil
}

func (s *CheckoutService) acceptMidtransCardTransaction(ctx context.Context, record *models.CheckoutRecord, gateway *BusinessGatewayRuntime, transaction *provider.MidtransCardTransaction) error {
	if transaction == nil {
		return fmt.Errorf("missing Midtrans transaction")
	}
	amount, err := strconv.ParseFloat(transaction.GrossAmount, 64)
	// Some Core API responses omit merchant_id/currency. The request is already
	// authenticated with this tenant's Server Key; verify these fields when sent.
	if err != nil || amount != record.InvoiceAmount || transaction.OrderID != record.ReferenceID ||
		(transaction.MerchantID != "" && transaction.MerchantID != gateway.Config.MidtransMerchantID) ||
		transaction.TransactionID == "" || (transaction.Currency != "" && transaction.Currency != "IDR") || transaction.PaymentType != "credit_card" {
		return fmt.Errorf("Midtrans transaction does not match checkout")
	}
	if transaction.RedirectURL != "" && !provider.ValidMidtransAuthenticationURL(transaction.RedirectURL) {
		return fmt.Errorf("invalid Midtrans authentication URL")
	}
	if transaction.TransactionStatus == "pending" && transaction.RedirectURL == "" && record.AuthenticationURL == nil {
		return fmt.Errorf("Midtrans authentication URL missing")
	}
	if _, err = s.db.ExecContext(ctx, `UPDATE checkout_sessions SET authentication_url=COALESCE(NULLIF($2,''),authentication_url) WHERE id=$1`, record.ID, transaction.RedirectURL); err != nil {
		return err
	}
	if err = s.repo.Activate(ctx, record.ID, record.InvoiceID, "midtrans", record.InvoiceAmount, record.ReferenceID, record.ExpiresAt); err != nil {
		return err
	}
	_, err = s.applyMidtransNotification(ctx, gateway, midtransNotification{
		OrderID: transaction.OrderID, TransactionID: transaction.TransactionID, MerchantID: transaction.MerchantID,
		StatusCode: transaction.StatusCode, GrossAmount: transaction.GrossAmount, TransactionStatus: transaction.TransactionStatus, FraudStatus: transaction.FraudStatus,
	})
	return err
}
