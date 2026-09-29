package models

import "time"

type CheckoutSession struct {
	ID                int64     `db:"id" json:"id"`
	InvoiceID         int       `db:"invoice_id" json:"invoice_id"`
	PaymentID         *int      `db:"payment_id" json:"payment_id,omitempty"`
	PublicTokenHash   string    `db:"public_token_hash" json:"-"`
	ReferenceID       string    `db:"reference_id" json:"reference_id"`
	Provider          string    `db:"provider" json:"provider"`
	ProviderSessionID *string   `db:"provider_session_id" json:"provider_session_id,omitempty"`
	PaymentLinkURL    *string   `db:"payment_link_url" json:"payment_link_url,omitempty"`
	Status            string    `db:"status" json:"status"`
	CustomerEmail     *string   `db:"customer_email" json:"customer_email,omitempty"`
	CustomerMobile    *string   `db:"customer_mobile" json:"customer_mobile,omitempty"`
	ExpiresAt         time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

type CheckoutRecord struct {
	ID                int64     `db:"checkout_id"`
	InvoiceID         int       `db:"invoice_id"`
	PaymentID         *int      `db:"payment_id"`
	ReferenceID       string    `db:"reference_id"`
	ProviderSessionID *string   `db:"provider_session_id"`
	PaymentLinkURL    *string   `db:"payment_link_url"`
	Status            string    `db:"checkout_status"`
	ExpiresAt         time.Time `db:"expires_at"`
	InvoiceNumber     string    `db:"invoice_number" json:"invoice_number"`
	InvoiceAmount     float64   `db:"invoice_amount" json:"amount"`
	InvoiceStatus     string    `db:"invoice_status" json:"invoice_status"`
	CustomerID        int       `db:"customer_id" json:"-"`
	CustomerName      string    `db:"customer_name" json:"customer_name"`
	DefaultEmail      *string   `db:"default_email" json:"customer_email,omitempty"`
	DefaultMobile     *string   `db:"default_mobile" json:"customer_mobile,omitempty"`
	BusinessName      string    `db:"business_name" json:"business_name"`
	PlanName          string    `db:"plan_name" json:"plan_name"`
}

type CheckoutLinkResponse struct {
	Token       string    `json:"token"`
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type PublicCheckoutResponse struct {
	InvoiceNumber  string    `json:"invoice_number"`
	Amount         float64   `json:"amount"`
	InvoiceStatus  string    `json:"invoice_status"`
	CheckoutStatus string    `json:"checkout_status"`
	CustomerName   string    `json:"customer_name"`
	CustomerEmail  *string   `json:"customer_email,omitempty"`
	CustomerMobile *string   `json:"customer_mobile,omitempty"`
	BusinessName   string    `json:"business_name"`
	PlanName       string    `json:"plan_name"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type StartCheckoutRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Mobile       string `json:"mobile"`
	ParentOrigin string `json:"parent_origin,omitempty"`
}

type StartCheckoutResponse struct {
	PaymentSessionID string    `json:"payment_session_id"`
	Mode             string    `json:"mode"`
	ComponentsSDKKey string    `json:"components_sdk_key"`
	ExpiresAt        time.Time `json:"expires_at"`
	Status           string    `json:"status"`
}
