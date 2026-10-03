package payment

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxMidtransResponseBytes = 1 << 20

type MidtransProvider struct {
	ServerKey string
	ClientKey string
	BaseURL   string
	Mock      bool
	client    *http.Client
}

type MidtransSnapRequest struct {
	OrderID   string
	Amount    int64
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type MidtransAPIError struct {
	StatusCode int
	Message    string
}

func (e *MidtransAPIError) Error() string {
	return fmt.Sprintf("Midtrans API returned HTTP %d: %s", e.StatusCode, e.Message)
}

func NewMidtransProvider(serverKey, clientKey, baseURL string, mock bool) *MidtransProvider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://app.sandbox.midtrans.com"
	}
	return &MidtransProvider{ServerKey: strings.TrimSpace(serverKey), ClientKey: strings.TrimSpace(clientKey),
		BaseURL: baseURL, Mock: mock, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *MidtransProvider) CreateSnapTransaction(ctx context.Context, request MidtransSnapRequest) (*MidtransSnapResponse, error) {
	if request.OrderID == "" || len(request.OrderID) > 50 || request.Amount < 1 || request.Email == "" {
		return nil, fmt.Errorf("Midtrans order ID, positive amount, and customer email are required")
	}
	for _, char := range request.OrderID {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("-_~.", char)) {
			return nil, fmt.Errorf("Midtrans order ID contains unsupported characters")
		}
	}
	if p.Mock {
		token, err := randomToken(24)
		if err != nil {
			return nil, err
		}
		return &MidtransSnapResponse{Token: "mock-" + token}, nil
	}
	if p.ServerKey == "" {
		return nil, fmt.Errorf("Midtrans Server Key is required")
	}
	body := map[string]any{
		"transaction_details": map[string]any{"order_id": request.OrderID, "gross_amount": request.Amount},
		"credit_card":         map[string]any{"secure": true},
		"customer_details": map[string]string{
			"first_name": request.FirstName, "last_name": request.LastName,
			"email": request.Email, "phone": request.Phone,
		},
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/snap/v1/transactions", bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(p.ServerKey, "")
	response, err := p.client.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("send Midtrans Snap request: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxMidtransResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(responseBody) > maxMidtransResponseBytes {
		return nil, fmt.Errorf("Midtrans response exceeds size limit")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiError struct {
			ErrorMessages []string `json:"error_messages"`
			Message       string   `json:"message"`
		}
		_ = json.Unmarshal(responseBody, &apiError)
		message := apiError.Message
		if message == "" && len(apiError.ErrorMessages) > 0 {
			message = strings.Join(apiError.ErrorMessages, "; ")
		}
		if message == "" {
			message = http.StatusText(response.StatusCode)
		}
		return nil, &MidtransAPIError{StatusCode: response.StatusCode, Message: message}
	}
	var result MidtransSnapResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("decode Midtrans Snap response: %w", err)
	}
	if strings.TrimSpace(result.Token) == "" {
		return nil, fmt.Errorf("Midtrans response did not include a Snap token")
	}
	if result.RedirectURL != "" {
		parsed, err := url.Parse(result.RedirectURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return nil, fmt.Errorf("Midtrans returned an invalid redirect URL")
		}
	}
	return &result, nil
}

func (p *MidtransProvider) Charge(req ChargeRequest) (*ChargeResponse, error) {
	if p.Mock {
		return p.mockCharge(req)
	}
	return nil, fmt.Errorf("Midtrans recurring charge is not implemented")
}

func (p *MidtransProvider) mockCharge(req ChargeRequest) (*ChargeResponse, error) {
	if req.GrossAmount <= 0 {
		return &ChargeResponse{Status: "failed", Message: "invalid amount"}, nil
	}
	return &ChargeResponse{ProviderTransactionID: "mock_midtrans_" + randomHex(16), Status: "success", Message: "Mock charge success"}, nil
}

// VerifyNotificationSignature implements Midtrans's documented SHA-512
// signature over order_id + status_code + gross_amount + ServerKey.
func VerifyMidtransNotificationSignature(orderID, statusCode, grossAmount, serverKey, received string) bool {
	if orderID == "" || statusCode == "" || grossAmount == "" || serverKey == "" || len(received) != sha512.Size*2 {
		return false
	}
	provided, err := hex.DecodeString(received)
	if err != nil {
		return false
	}
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	return subtle.ConstantTimeCompare(sum[:], provided) == 1
}

func randomToken(n int) (string, error) {
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func randomHex(n int) string {
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		return "unavailable"
	}
	return hex.EncodeToString(data)
}
