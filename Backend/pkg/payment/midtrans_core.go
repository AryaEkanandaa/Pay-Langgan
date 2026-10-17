package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type MidtransCardRequest struct {
	OrderID string
	Amount  int64
	TokenID string
	Name    string
	Email   string
	Phone   string
}

type MidtransCardTransaction struct {
	StatusCode        string `json:"status_code"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
	PaymentType       string `json:"payment_type"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	MerchantID        string `json:"merchant_id"`
	RedirectURL       string `json:"redirect_url"`
}

// BaseURL must be the Core API origin, not the Snap application origin.
func (p *MidtransProvider) ChargeCard(ctx context.Context, request MidtransCardRequest) (*MidtransCardTransaction, error) {
	if request.TokenID == "" || len(request.TokenID) > 255 || request.Amount < 1 || request.OrderID == "" || len(request.OrderID) > 50 {
		return nil, fmt.Errorf("invalid Midtrans card transaction")
	}
	body := map[string]any{
		"payment_type":        "credit_card",
		"transaction_details": map[string]any{"order_id": request.OrderID, "gross_amount": request.Amount},
		"credit_card":         map[string]any{"token_id": request.TokenID, "authentication": true},
		"customer_details":    map[string]string{"first_name": request.Name, "email": request.Email, "phone": request.Phone},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return p.cardRequest(ctx, http.MethodPost, "/v2/charge", raw)
}

func (p *MidtransProvider) GetCardTransaction(ctx context.Context, orderID string) (*MidtransCardTransaction, error) {
	return p.cardRequest(ctx, http.MethodGet, "/v2/"+url.PathEscape(orderID)+"/status", nil)
}

func (p *MidtransProvider) cardRequest(ctx context.Context, method, path string, body []byte) (*MidtransCardTransaction, error) {
	request, err := http.NewRequestWithContext(ctx, method, p.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(p.ServerKey, "")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Midtrans card request: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxMidtransResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxMidtransResponseBytes {
		return nil, fmt.Errorf("Midtrans response exceeds size limit")
	}
	var result MidtransCardTransaction
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("invalid Midtrans card response")
	}
	code, _ := strconv.Atoi(result.StatusCode)
	if code == 0 {
		code = response.StatusCode
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || code < 200 || code >= 300 {
		return nil, &MidtransAPIError{StatusCode: code, Message: "card transaction request rejected"}
	}
	if result.RedirectURL != "" && !ValidMidtransAuthenticationURL(result.RedirectURL) {
		return nil, fmt.Errorf("invalid Midtrans authentication URL")
	}
	return &result, nil
}

func ValidMidtransAuthenticationURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "card.midtrans.com" || host == "api.midtrans.com" || host == "api.sandbox.midtrans.com" || host == "api.veritrans.co.id" || host == "api.sandbox.veritrans.co.id"
}
