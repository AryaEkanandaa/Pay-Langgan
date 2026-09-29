package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RecurringChargeRequest struct {
	ReferenceID      string
	PaymentTokenID   string
	Amount           int64
	ReturnURL        string
	FrequencyDays    int
	ConsentExpiresAt time.Time
}

type RecurringChargeResponse struct {
	PaymentRequestID string `json:"payment_request_id"`
	ReferenceID      string `json:"reference_id"`
	Currency         string `json:"currency"`
	Amount           int64  `json:"request_amount"`
	Status           string `json:"status"`
	FailureCode      string `json:"failure_code"`
	PaymentID        string `json:"latest_payment_id"`
}

func (p *XenditProvider) CreateRecurringCharge(ctx context.Context, r RecurringChargeRequest) (*RecurringChargeResponse, error) {
	u, err := url.Parse(r.ReturnURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || r.ReferenceID == "" || len(r.ReferenceID) > 64 || r.PaymentTokenID == "" ||
		r.Amount < 5000 || r.Amount > 200000000 || r.FrequencyDays < 1 || !r.ConsentExpiresAt.After(time.Now()) {
		return nil, &ValidationError{Err: fmt.Errorf("invalid recurring charge request")}
	}
	if p.mock {
		return &RecurringChargeResponse{PaymentRequestID: "pr-mock-" + r.ReferenceID, ReferenceID: r.ReferenceID, Currency: "IDR", Amount: r.Amount, Status: "SUCCEEDED", PaymentID: "py-mock-" + r.ReferenceID}, nil
	}
	body, err := json.Marshal(map[string]any{
		"reference_id": r.ReferenceID, "payment_token_id": r.PaymentTokenID, "type": "PAY", "country": "ID", "currency": "IDR",
		"request_amount": r.Amount, "capture_method": "AUTOMATIC",
		"channel_properties": map[string]any{"card_on_file_type": "RECURRING", "transaction_sequence": "SUBSEQUENT",
			"success_return_url": r.ReturnURL, "failure_return_url": r.ReturnURL,
			"recurring_configuration": map[string]any{"recurring_frequency": r.FrequencyDays, "recurring_expiry": r.ConsentExpiresAt.UTC().Format("2006-01-02")}},
	})
	if err != nil {
		return nil, err
	}
	var response RecurringChargeResponse
	if err := p.requestJSON(ctx, http.MethodPost, "/v3/payment_requests", string(body), &response); err != nil {
		return nil, err
	}
	if response.PaymentRequestID == "" || response.ReferenceID != r.ReferenceID {
		return nil, fmt.Errorf("incomplete recurring charge response")
	}
	return &response, nil
}

func (p *XenditProvider) GetRecurringCharge(ctx context.Context, id string) (*RecurringChargeResponse, error) {
	if id == "" || strings.ContainsAny(id, "/?#") {
		return nil, fmt.Errorf("invalid payment request ID")
	}
	if p.mock {
		return nil, fmt.Errorf("mock recurring results are only available on creation")
	}
	var response RecurringChargeResponse
	if err := p.requestJSON(ctx, http.MethodGet, "/v3/payment_requests/"+url.PathEscape(id), "", &response); err != nil {
		return nil, err
	}
	if response.PaymentRequestID != id {
		return nil, fmt.Errorf("mismatched recurring charge response")
	}
	return &response, nil
}
