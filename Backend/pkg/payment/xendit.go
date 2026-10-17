package payment

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"time"
)

const maxXenditResponseBytes = 1 << 20

type XenditProvider struct {
	secretKey     string
	callbackToken string
	baseURL       string
	mock          bool
	httpClient    *http.Client
	mockMu        sync.Mutex
	mockSessions  map[string]XenditSessionResponse
}

type XenditSessionRequest struct {
	SessionType         string // Empty means PAY; SAVE enrolls a card without charging.
	ReferenceID         string
	Amount              int64
	CustomerReferenceID string
	CustomerName        string
	CustomerEmail       string
	CustomerMobile      string
	Description         string
	ComponentsOrigins   []string
	ReturnURL           string
	RecurringFrequency  int
	RecurringExpiry     string
}

type ComponentsConfiguration struct {
	Origins   []string `json:"origins"`
	ReturnURL string   `json:"return_url"`
}

type XenditSessionResponse struct {
	SessionType      string    `json:"session_type"`
	Amount           int64     `json:"amount"`
	Currency         string    `json:"currency"`
	PaymentSessionID string    `json:"payment_session_id"`
	ReferenceID      string    `json:"reference_id"`
	Status           string    `json:"status"`
	PaymentLinkURL   string    `json:"payment_link_url"`
	Mode             string    `json:"mode"`
	ComponentsSDKKey string    `json:"components_sdk_key"`
	ExpiresAt        time.Time `json:"expires_at"`
	PaymentTokenID   string    `json:"payment_token_id"`
	PaymentRequestID string    `json:"payment_request_id"`
	PaymentID        string    `json:"payment_id"`
}

type ValidationError struct{ Err error }

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error { return e.Err }

// Only these outcomes prove the creation request did not create a session.
// Network errors, 408/409/429/5xx and malformed responses remain uncertain.
func SessionCreationRejected(err error) bool {
	var validation *ValidationError
	if errors.As(err, &validation) {
		return true
	}
	var api *XenditAPIError
	if errors.As(err, &api) {
		return api.StatusCode == 400 || api.StatusCode == 401 || api.StatusCode == 403 || api.StatusCode == 404 || api.StatusCode == 422 ||
			(api.StatusCode == 409 && api.Code == "DUPLICATE_ERROR" && strings.Contains(strings.ToLower(api.Message), "customer"))
	}
	var midtrans *MidtransAPIError
	return errors.As(err, &midtrans) && (midtrans.StatusCode == 400 || midtrans.StatusCode == 401 || midtrans.StatusCode == 403 || midtrans.StatusCode == 404 || midtrans.StatusCode == 422)
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
		mockSessions:  make(map[string]XenditSessionResponse),
	}, nil
}

func (p *XenditProvider) CreateCardSession(ctx context.Context, request XenditSessionRequest) (*XenditSessionResponse, error) {
	if err := validateXenditSessionRequest(request); err != nil {
		return nil, &ValidationError{Err: err}
	}
	if p.mock {
		id, err := secureHex(13)
		if err != nil {
			return nil, fmt.Errorf("generate mock Xendit session ID: %w", err)
		}
		sessionID := "ps-" + id
		sessionType := request.SessionType
		if sessionType == "" {
			sessionType = "PAY"
		}
		session := XenditSessionResponse{
			SessionType: sessionType,
			Amount:      request.Amount, Currency: "IDR",
			PaymentSessionID: sessionID,
			ReferenceID:      request.ReferenceID,
			Status:           "ACTIVE",
			Mode:             "COMPONENTS",
			ComponentsSDKKey: "mock-components-" + sessionID,
			ExpiresAt:        time.Now().UTC().Add(30 * time.Minute),
		}
		stored := session
		stored.ComponentsSDKKey = ""
		p.mockMu.Lock()
		p.mockSessions[sessionID] = stored
		p.mockMu.Unlock()
		return &session, nil
	}

	payload := struct {
		ReferenceID             string                  `json:"reference_id"`
		SessionType             string                  `json:"session_type"`
		Mode                    string                  `json:"mode"`
		Amount                  int64                   `json:"amount"`
		Currency                string                  `json:"currency"`
		Country                 string                  `json:"country"`
		CaptureMethod           string                  `json:"capture_method,omitempty"`
		AllowedPaymentChannels  []string                `json:"allowed_payment_channels"`
		AllowSavePaymentMethod  string                  `json:"allow_save_payment_method,omitempty"`
		Locale                  string                  `json:"locale"`
		Description             string                  `json:"description"`
		ComponentsConfiguration ComponentsConfiguration `json:"components_configuration"`
		ChannelProperties       map[string]any          `json:"channel_properties,omitempty"`
		Customer                struct {
			ReferenceID      string `json:"reference_id"`
			Type             string `json:"type"`
			Email            string `json:"email"`
			MobileNumber     string `json:"mobile_number"`
			IndividualDetail struct {
				GivenNames string `json:"given_names"`
			} `json:"individual_detail"`
		} `json:"customer"`
	}{
		ReferenceID:             request.ReferenceID,
		SessionType:             "PAY",
		Mode:                    "COMPONENTS",
		Amount:                  request.Amount,
		Currency:                "IDR",
		Country:                 "ID",
		CaptureMethod:           "AUTOMATIC",
		AllowedPaymentChannels:  []string{"CARDS"},
		AllowSavePaymentMethod:  "FORCED",
		Locale:                  "id",
		Description:             request.Description,
		ComponentsConfiguration: ComponentsConfiguration{Origins: request.ComponentsOrigins, ReturnURL: request.ReturnURL},
	}
	if request.SessionType == "SAVE" {
		payload.SessionType = "SAVE"
		payload.CaptureMethod = ""
		payload.AllowSavePaymentMethod = ""
		if request.RecurringFrequency > 0 && request.RecurringExpiry != "" {
			payload.ChannelProperties = map[string]any{"cards": map[string]any{
				"card_on_file_type":       "RECURRING",
				"recurring_configuration": map[string]any{"recurring_frequency": request.RecurringFrequency, "recurring_expiry": request.RecurringExpiry},
			}}
		}
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
	session, err := p.requestSession(ctx, http.MethodPost, "/sessions", string(body))
	var duplicate *XenditAPIError
	if errors.As(err, &duplicate) && duplicate.StatusCode == http.StatusConflict && duplicate.Code == "DUPLICATE_ERROR" && strings.Contains(strings.ToLower(duplicate.Message), "customer") {
		// The customer object creates a customer; repeat purchases must use
		// customer_id. Look up only within the account authenticated by this key.
		var customers struct {
			Data []struct {
				ID          string `json:"id"`
				ReferenceID string `json:"reference_id"`
			} `json:"data"`
		}
		lookupErr := p.requestJSON(ctx, http.MethodGet, "/customers?reference_id="+url.QueryEscape(request.CustomerReferenceID), "", &customers)
		if lookupErr == nil && len(customers.Data) == 1 && customers.Data[0].ReferenceID == request.CustomerReferenceID && strings.HasPrefix(customers.Data[0].ID, "cust-") {
			var reuse map[string]any
			if json.Unmarshal(body, &reuse) == nil {
				delete(reuse, "customer")
				reuse["customer_id"] = customers.Data[0].ID
				retryBody, encodeErr := json.Marshal(reuse)
				if encodeErr != nil {
					return nil, encodeErr
				}
				session, err = p.requestSession(ctx, http.MethodPost, "/sessions", string(retryBody))
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if session.ReferenceID != request.ReferenceID || !activeComponentsSession(session) {
		return nil, fmt.Errorf("Xendit returned an incomplete components session")
	}
	return session, nil
}

// GetCardSession retrieves the client key without persisting it locally.
func (p *XenditProvider) GetCardSession(ctx context.Context, sessionID, referenceID string) (*XenditSessionResponse, error) {
	if sessionID == "" || strings.ContainsAny(sessionID, "/?#") {
		return nil, fmt.Errorf("invalid Xendit session ID")
	}
	if p.mock {
		p.mockMu.Lock()
		session, ok := p.mockSessions[sessionID]
		p.mockMu.Unlock()
		if !ok {
			return nil, &XenditAPIError{StatusCode: 404, Code: "NOT_FOUND", Message: "mock session not found in this process"}
		}
		if session.ReferenceID != referenceID {
			return nil, fmt.Errorf("Xendit returned a mismatched session")
		}
		if session.ExpiresAt.After(time.Now()) {
			session.ComponentsSDKKey = "mock-components-" + sessionID
		} else {
			session.Status = "EXPIRED"
		}
		return &session, nil
	}
	session, err := p.requestSession(ctx, http.MethodGet, "/sessions/"+url.PathEscape(sessionID), "")
	if err != nil {
		return nil, err
	}
	if session.PaymentSessionID != sessionID || session.ReferenceID != referenceID {
		return nil, fmt.Errorf("Xendit returned a mismatched session")
	}
	return session, nil
}

// CancelCardSession invalidates an ACTIVE PAY or SAVE session. The caller must
// still decide whether its own invoice or trial enrollment can be cancelled.
func (p *XenditProvider) CancelCardSession(ctx context.Context, sessionID, referenceID string) (*XenditSessionResponse, error) {
	if sessionID == "" || strings.ContainsAny(sessionID, "/?#") || referenceID == "" {
		return nil, fmt.Errorf("invalid Xendit session identity")
	}
	if p.mock {
		p.mockMu.Lock()
		defer p.mockMu.Unlock()
		session, ok := p.mockSessions[sessionID]
		if !ok || session.ReferenceID != referenceID {
			return nil, &XenditAPIError{StatusCode: 404, Code: "NOT_FOUND", Message: "mock session not found"}
		}
		if session.Status != "ACTIVE" {
			return nil, &XenditAPIError{StatusCode: 422, Code: "INVALID_SESSION_STATUS", Message: "session is not active"}
		}
		session.Status = "CANCELED"
		p.mockSessions[sessionID] = session
		return &session, nil
	}
	session, err := p.requestSession(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/cancel", "")
	if err != nil {
		return nil, err
	}
	if session.PaymentSessionID != sessionID || session.ReferenceID != referenceID || session.Status != "CANCELED" {
		return nil, fmt.Errorf("Xendit returned a mismatched cancellation")
	}
	return session, nil
}

func activeComponentsSession(session *XenditSessionResponse) bool {
	return session != nil && session.PaymentSessionID != "" && session.Mode == "COMPONENTS" &&
		session.Status == "ACTIVE" && strings.TrimSpace(session.ComponentsSDKKey) != "" && session.ExpiresAt.After(time.Now())
}

func (p *XenditProvider) requestSession(ctx context.Context, method, path, body string) (*XenditSessionResponse, error) {
	var session XenditSessionResponse
	if err := p.requestJSON(ctx, method, path, body, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (p *XenditProvider) requestJSON(ctx context.Context, method, path, body string, target any) error {
	httpRequest, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Xendit request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	if strings.HasPrefix(path, "/v3/") {
		httpRequest.Header.Set("api-version", "2024-11-11")
	}
	if strings.HasPrefix(path, "/customers?") {
		httpRequest.Header.Set("api-version", "2020-10-31")
	}
	httpRequest.SetBasicAuth(p.secretKey, "")

	response, err := p.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("send Xendit request: %w", err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, maxXenditResponseBytes+1)
	responseBody, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("read Xendit response: %w", err)
	}
	if len(responseBody) > maxXenditResponseBytes {
		return fmt.Errorf("Xendit response exceeds size limit")
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
		return &XenditAPIError{StatusCode: response.StatusCode, Code: apiError.ErrorCode, Message: apiError.Message}
	}
	if err := json.Unmarshal(responseBody, target); err != nil {
		return fmt.Errorf("decode Xendit response: %w", err)
	}
	return nil
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
	if request.SessionType != "" && request.SessionType != "PAY" && request.SessionType != "SAVE" {
		return fmt.Errorf("unsupported session type")
	}
	if request.SessionType == "SAVE" && request.Amount != 0 {
		return fmt.Errorf("card registration must not charge an amount")
	}
	if request.SessionType == "SAVE" && ((request.RecurringFrequency == 0) != (request.RecurringExpiry == "") || request.RecurringFrequency < 0) {
		return fmt.Errorf("invalid recurring configuration for SAVE session")
	}
	if request.RecurringExpiry != "" {
		if _, err := time.Parse("2006-01-02", request.RecurringExpiry); err != nil {
			return fmt.Errorf("recurring expiry must use YYYY-MM-DD")
		}
	}
	if request.SessionType != "SAVE" && (request.Amount < 5000 || request.Amount > 200000000) {
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
	if len(request.ComponentsOrigins) == 0 {
		return fmt.Errorf("components origins are required")
	}
	for _, origin := range request.ComponentsOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.ForceQuery || strings.Contains(origin, "*") ||
			(parsed.Scheme != "https" && !(parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1"))) {
			return fmt.Errorf("components origin must be HTTPS or a local development HTTP origin without a path")
		}
	}
	for name, value := range map[string]string{"components return URL": request.ReturnURL} {
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
