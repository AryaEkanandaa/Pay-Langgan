package payment

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

const maxXenditResponseBytes = 1 << 20

type XenditProvider struct {
	secretKey     string
	callbackToken string
	baseURL       string
	mock          bool
	httpClient    *http.Client
}

type XenditSessionRequest struct {
	ReferenceID         string
	Amount              int64
	CustomerReferenceID string
	CustomerName        string
	CustomerEmail       string
	CustomerMobile      string
	Description         string
	SuccessReturnURL    string
	CancelReturnURL     string
}

type XenditSessionResponse struct {
	PaymentSessionID string `json:"payment_session_id"`
	ReferenceID      string `json:"reference_id"`
	Status           string `json:"status"`
	PaymentLinkURL   string `json:"payment_link_url"`
	PaymentTokenID   string `json:"payment_token_id"`
	PaymentRequestID string `json:"payment_request_id"`
	PaymentID        string `json:"payment_id"`
}

type XenditAPIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *XenditAPIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("xendit API returned HTTP %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("xendit API returned HTTP %d: %s", e.StatusCode, e.Message)
}

func NewXenditProvider(secretKey, callbackToken, baseURL string, mock bool) (*XenditProvider, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.xendit.co"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid Xendit base URL")
	}
	if !mock && parsed.Scheme != "https" {
		return nil, fmt.Errorf("Xendit base URL must use HTTPS")
	}
	if !mock && strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("Xendit secret key is required")
	}
	if !mock && strings.TrimSpace(callbackToken) == "" {
		return nil, fmt.Errorf("Xendit callback token is required")
	}
	return &XenditProvider{
		secretKey:     secretKey,
		callbackToken: callbackToken,
		baseURL:       baseURL,
		mock:          mock,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (p *XenditProvider) CreateCardSession(ctx context.Context, request XenditSessionRequest) (*XenditSessionResponse, error) {
	if err := validateXenditSessionRequest(request); err != nil {
		return nil, err
	}
	if p.mock {
		id, err := secureHex(13)
		if err != nil {
			return nil, fmt.Errorf("generate mock Xendit session ID: %w", err)
		}
		sessionID := "ps-" + id
		return &XenditSessionResponse{
			PaymentSessionID: sessionID,
			ReferenceID:      request.ReferenceID,
			Status:           "ACTIVE",
			PaymentLinkURL:   "https://checkout.xendit.test/" + sessionID,
		}, nil
	}

	payload := struct {
		ReferenceID            string   `json:"reference_id"`
		SessionType            string   `json:"session_type"`
		Mode                   string   `json:"mode"`
		Amount                 int64    `json:"amount"`
		Currency               string   `json:"currency"`
		Country                string   `json:"country"`
		CaptureMethod          string   `json:"capture_method"`
		AllowedPaymentChannels []string `json:"allowed_payment_channels"`
		AllowSavePaymentMethod string   `json:"allow_save_payment_method"`
		Locale                 string   `json:"locale"`
		Description            string   `json:"description"`
		SuccessReturnURL       string   `json:"success_return_url"`
		CancelReturnURL        string   `json:"cancel_return_url"`
		Customer               struct {
			ReferenceID      string `json:"reference_id"`
			Type             string `json:"type"`
			Email            string `json:"email"`
			MobileNumber     string `json:"mobile_number"`
			IndividualDetail struct {
				GivenNames string `json:"given_names"`
			} `json:"individual_detail"`
		} `json:"customer"`
	}{
		ReferenceID:            request.ReferenceID,
		SessionType:            "PAY",
		Mode:                   "PAYMENT_LINK",
		Amount:                 request.Amount,
		Currency:               "IDR",
		Country:                "ID",
		CaptureMethod:          "AUTOMATIC",
		AllowedPaymentChannels: []string{"CARDS"},
		AllowSavePaymentMethod: "FORCED",
		Locale:                 "id",
		Description:            request.Description,
		SuccessReturnURL:       request.SuccessReturnURL,
		CancelReturnURL:        request.CancelReturnURL,
	}
	payload.Customer.ReferenceID = request.CustomerReferenceID
	payload.Customer.Type = "INDIVIDUAL"
	payload.Customer.Email = request.CustomerEmail
	payload.Customer.MobileNumber = request.CustomerMobile
	payload.Customer.IndividualDetail.GivenNames = request.CustomerName

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Xendit session: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/sessions", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create Xendit request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.SetBasicAuth(p.secretKey, "")

	response, err := p.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("send Xendit session request: %w", err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, maxXenditResponseBytes+1)
	responseBody, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read Xendit response: %w", err)
	}
	if len(responseBody) > maxXenditResponseBytes {
		return nil, fmt.Errorf("Xendit response exceeds size limit")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiError struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}
		_ = json.Unmarshal(responseBody, &apiError)
		if apiError.Message == "" {
			apiError.Message = http.StatusText(response.StatusCode)
		}
		return nil, &XenditAPIError{StatusCode: response.StatusCode, Code: apiError.ErrorCode, Message: apiError.Message}
	}
	var session XenditSessionResponse
	if err := json.Unmarshal(responseBody, &session); err != nil {
		return nil, fmt.Errorf("decode Xendit session response: %w", err)
	}
	if session.PaymentSessionID == "" || session.ReferenceID != request.ReferenceID || session.PaymentLinkURL == "" {
		return nil, fmt.Errorf("Xendit returned an incomplete session")
	}
	return &session, nil
}

func (p *XenditProvider) VerifyCallbackToken(received string) bool {
	expected := []byte(p.callbackToken)
	actual := []byte(received)
	if len(expected) == 0 || len(expected) != len(actual) {
		return false
	}
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func validateXenditSessionRequest(request XenditSessionRequest) error {
	if strings.TrimSpace(request.ReferenceID) == "" || len(request.ReferenceID) > 255 {
		return fmt.Errorf("Xendit reference ID is required and must be at most 255 characters")
	}
	if request.Amount < 5000 || request.Amount > 200000000 {
		return fmt.Errorf("card payment amount must be between IDR 5,000 and IDR 200,000,000")
	}
	if strings.TrimSpace(request.CustomerReferenceID) == "" || strings.TrimSpace(request.CustomerName) == "" {
		return fmt.Errorf("customer reference and name are required")
	}
	address, err := mail.ParseAddress(request.CustomerEmail)
	if err != nil || address.Address != request.CustomerEmail {
		return fmt.Errorf("valid customer email is required")
	}
	if !validE164(request.CustomerMobile) {
		return fmt.Errorf("customer mobile number must use E.164 format")
	}
	for name, value := range map[string]string{"success return URL": request.SuccessReturnURL, "cancel return URL": request.CancelReturnURL} {
		parsed, err := url.ParseRequestURI(value)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must be an HTTPS URL", name)
		}
	}
	return nil
}

func validE164(value string) bool {
	if len(value) < 9 || len(value) > 16 || value[0] != '+' {
		return false
	}
	for _, char := range value[1:] {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func secureHex(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
