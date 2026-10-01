package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

type HostedSessionRequest struct {
	ReferenceID       string
	Amount            int64
	BusinessID        string
	PayerName         string
	PayerEmail        string
	SuccessURL        string
	CancelURL         string
	SavePaymentMethod bool
}

// PlatformComponentsRequest charges the business and saves its card for renewals.
type PlatformComponentsRequest struct {
	ReferenceID string
	Amount      int64
	BusinessID  string
	PayerName   string
	PayerEmail  string
	Origin      string
	ReturnURL   string
}

func (p *XenditProvider) CreatePlatformComponentsSession(ctx context.Context, r PlatformComponentsRequest) (*XenditSessionResponse, error) {
	if r.ReferenceID == "" || r.BusinessID == "" || strings.TrimSpace(r.PayerName) == "" || r.Amount < 5000 || r.Amount > 200000000 {
		return nil, &ValidationError{Err: fmt.Errorf("invalid platform components request")}
	}
	address, err := mail.ParseAddress(r.PayerEmail)
	if err != nil || address.Address != r.PayerEmail {
		return nil, &ValidationError{Err: fmt.Errorf("valid payer email is required")}
	}
	origin, err := url.Parse(r.Origin)
	if err != nil || origin.Scheme != "https" || origin.Host == "" || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return nil, &ValidationError{Err: fmt.Errorf("platform components origin must be an HTTPS origin")}
	}
	returnURL, err := url.Parse(r.ReturnURL)
	if err != nil || returnURL.Scheme != "https" || returnURL.Host != origin.Host || returnURL.User != nil {
		return nil, &ValidationError{Err: fmt.Errorf("platform return URL must use the checkout origin")}
	}
	if p.mock {
		id, err := secureHex(13)
		if err != nil {
			return nil, err
		}
		sessionID := "ps-" + id
		session := XenditSessionResponse{PaymentSessionID: sessionID, ReferenceID: r.ReferenceID, Amount: r.Amount, Currency: "IDR", Mode: "COMPONENTS", Status: "ACTIVE", ComponentsSDKKey: "mock-components-" + sessionID, ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}
		stored := session
		stored.ComponentsSDKKey = ""
		p.mockMu.Lock()
		p.mockSessions[sessionID] = stored
		p.mockMu.Unlock()
		return &session, nil
	}
	body := map[string]any{
		"reference_id": r.ReferenceID, "session_type": "PAY", "mode": "COMPONENTS",
		"amount": r.Amount, "currency": "IDR", "country": "ID", "capture_method": "AUTOMATIC",
		"allowed_payment_channels": []string{"CARDS"}, "allow_save_payment_method": "FORCED",
		"description": "Langganan PayLanggan bulanan", "locale": "id",
		"components_configuration": ComponentsConfiguration{Origins: []string{r.Origin}, ReturnURL: r.ReturnURL},
		"customer":                 map[string]any{"reference_id": r.ReferenceID, "type": "INDIVIDUAL", "email": r.PayerEmail, "individual_detail": map[string]any{"given_names": r.PayerName}},
		"channel_properties": map[string]any{"cards": map[string]any{
			"card_on_file_type": "RECURRING", "recurring_configuration": map[string]any{
				"recurring_frequency": 30, "recurring_expiry": time.Now().UTC().AddDate(1, 0, 0).Format("2006-01-02"),
			},
		}},
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	session, err := p.requestSession(ctx, http.MethodPost, "/sessions", string(encoded))
	if err != nil {
		return nil, err
	}
	if session.ReferenceID != r.ReferenceID || session.Amount != r.Amount || session.Currency != "IDR" || !activeComponentsSession(session) {
		return nil, fmt.Errorf("Xendit returned an incomplete platform components session")
	}
	return session, nil
}

func (p *XenditProvider) CreateHostedSession(ctx context.Context, r HostedSessionRequest) (*XenditSessionResponse, error) {
	if r.ReferenceID == "" || r.BusinessID == "" || strings.TrimSpace(r.PayerName) == "" || r.Amount < 5000 || r.Amount > 200000000 {
		return nil, &ValidationError{Err: fmt.Errorf("invalid hosted session request")}
	}
	address, err := mail.ParseAddress(r.PayerEmail)
	if err != nil || address.Address != r.PayerEmail {
		return nil, &ValidationError{Err: fmt.Errorf("valid payer email is required")}
	}
	for _, value := range []string{r.SuccessURL, r.CancelURL} {
		u, err := url.Parse(value)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return nil, &ValidationError{Err: fmt.Errorf("hosted return URLs must use HTTPS")}
		}
	}
	if p.mock {
		id, err := secureHex(13)
		if err != nil {
			return nil, err
		}
		return &XenditSessionResponse{PaymentSessionID: "ps-" + id, ReferenceID: r.ReferenceID,
			Amount: r.Amount, Currency: "IDR", Mode: "PAYMENT_LINK", Status: "ACTIVE",
			PaymentLinkURL: "https://mock.xendit.invalid/" + id, ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}, nil
	}
	allowSave := "DISABLED"
	if r.SavePaymentMethod {
		allowSave = "FORCED"
	}
	body := map[string]any{
		"reference_id": r.ReferenceID, "session_type": "PAY", "mode": "PAYMENT_LINK",
		"amount": r.Amount, "currency": "IDR", "country": "ID", "capture_method": "AUTOMATIC",
		"allowed_payment_channels": []string{"CARDS"}, "allow_save_payment_method": allowSave,
		"description": "Langganan PayLanggan bulanan", "locale": "id",
		"success_return_url": r.SuccessURL, "cancel_return_url": r.CancelURL,
		"customer": map[string]any{"reference_id": r.ReferenceID, "type": "INDIVIDUAL",
			"email": r.PayerEmail, "individual_detail": map[string]any{"given_names": r.PayerName}},
	}
	if r.SavePaymentMethod {
		body["channel_properties"] = map[string]any{"cards": map[string]any{
			"card_on_file_type": "RECURRING", "recurring_configuration": map[string]any{
				"recurring_frequency": 30, "recurring_expiry": time.Now().UTC().AddDate(1, 0, 0).Format("2006-01-02"),
			},
		}}
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	session, err := p.requestSession(ctx, http.MethodPost, "/sessions", string(encoded))
	if err != nil {
		return nil, err
	}
	if session.ReferenceID != r.ReferenceID || session.Mode != "PAYMENT_LINK" || session.Status != "ACTIVE" ||
		session.PaymentSessionID == "" || session.PaymentLinkURL == "" || session.Amount != r.Amount || session.Currency != "IDR" {
		return nil, fmt.Errorf("Xendit returned an incomplete hosted session")
	}
	return session, nil
}

func (p *XenditProvider) GetHostedSession(ctx context.Context, sessionID string) (*XenditSessionResponse, error) {
	if sessionID == "" || strings.ContainsAny(sessionID, "/?#") {
		return nil, fmt.Errorf("invalid Xendit session ID")
	}
	if p.mock {
		return nil, fmt.Errorf("mock hosted sessions are not retrievable")
	}
	return p.requestSession(ctx, http.MethodGet, "/sessions/"+url.PathEscape(sessionID), "")
}
