package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CreateRecurringCharge uses a Midtrans saved_token_id for a merchant-initiated
// subsequent card charge. ReferenceID is the stable Midtrans order_id, allowing
// a charge with an uncertain HTTP response to be reconciled without retrying it.
func (p *MidtransProvider) CreateRecurringCharge(ctx context.Context, request RecurringChargeRequest) (*RecurringChargeResponse, error) {
	if request.ReferenceID == "" || len(request.ReferenceID) > 50 || request.PaymentTokenID == "" ||
		len(request.PaymentTokenID) > 255 || request.Amount < 5000 || request.Amount > 200000000 {
		return nil, &ValidationError{Err: fmt.Errorf("invalid Midtrans recurring charge request")}
	}
	for _, char := range request.ReferenceID {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("-_~.", char)) {
			return nil, &ValidationError{Err: fmt.Errorf("invalid Midtrans recurring order ID")}
		}
	}
	body := map[string]any{
		"payment_type": "credit_card",
		"transaction_details": map[string]any{
			"order_id": request.ReferenceID, "gross_amount": request.Amount,
		},
		"credit_card": map[string]any{"token_id": request.PaymentTokenID},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	transaction, err := p.cardRequest(ctx, "POST", "/v2/charge", raw)
	if err != nil {
		return nil, err
	}
	return midtransRecurringResponse(transaction, request.ReferenceID, request.Amount)
}

// GetRecurringCharge reconciles using Midtrans order_id, which is the stable
// reference created by PayLanggan before the charge request is sent.
func (p *MidtransProvider) GetRecurringCharge(ctx context.Context, orderID string) (*RecurringChargeResponse, error) {
	if orderID == "" || len(orderID) > 50 || strings.ContainsAny(orderID, "/?#") {
		return nil, &ValidationError{Err: fmt.Errorf("invalid Midtrans recurring order ID")}
	}
	transaction, err := p.GetCardTransaction(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return midtransRecurringResponse(transaction, orderID, 0)
}

func midtransRecurringResponse(transaction *MidtransCardTransaction, orderID string, expectedAmount int64) (*RecurringChargeResponse, error) {
	if transaction == nil || transaction.OrderID != orderID || transaction.TransactionID == "" || transaction.PaymentType != "credit_card" || transaction.Currency != "IDR" {
		return nil, fmt.Errorf("Midtrans recurring response identity mismatch")
	}
	amount, err := strconv.ParseFloat(transaction.GrossAmount, 64)
	if err != nil || amount < 0 || amount != float64(int64(amount)) || expectedAmount > 0 && int64(amount) != expectedAmount {
		return nil, fmt.Errorf("Midtrans recurring response amount mismatch")
	}
	status := "REQUIRES_ACTION"
	failureCode := ""
	switch transaction.TransactionStatus {
	case "settlement":
		status = "SUCCEEDED"
	case "capture":
		if transaction.FraudStatus == "accept" {
			status = "SUCCEEDED"
		} else if transaction.FraudStatus == "deny" {
			status, failureCode = "FAILED", "SUSPECTED_FRAUDULENT"
		}
	case "pending":
		status = "ACCEPTING_PAYMENTS"
	case "deny":
		status, failureCode = "FAILED", "CARD_DECLINED"
	case "expire":
		status, failureCode = "EXPIRED", "EXPIRED_CARD"
	case "cancel":
		status, failureCode = "CANCELED", "USER_DID_NOT_AUTHORIZE"
	default:
		return nil, fmt.Errorf("unsupported Midtrans recurring status")
	}
	return &RecurringChargeResponse{
		PaymentRequestID: orderID,
		ReferenceID:      transaction.OrderID,
		Currency:         "IDR",
		Amount:           int64(amount),
		Status:           status,
		FailureCode:      failureCode,
		PaymentID:        transaction.TransactionID,
	}, nil
}

var _ interface {
	CreateRecurringCharge(context.Context, RecurringChargeRequest) (*RecurringChargeResponse, error)
	GetRecurringCharge(context.Context, string) (*RecurringChargeResponse, error)
} = (*MidtransProvider)(nil)
