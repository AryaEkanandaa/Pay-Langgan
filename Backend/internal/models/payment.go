package models

import (
	"time"
)

type Payment struct {
	ID                    int        `db:"id" json:"id"`
	InvoiceID             int        `db:"invoice_id" json:"invoice_id"`
	Provider              string     `db:"provider" json:"provider"`
	ProviderTransactionID *string    `db:"provider_transaction_id" json:"provider_transaction_id"`
	Amount                float64    `db:"amount" json:"amount"`
	Status                string     `db:"status" json:"status"`
	PaidAt                *time.Time `db:"paid_at" json:"paid_at"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
}

type PaymentMethod struct {
	ID         int       `db:"id" json:"id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	Provider   string    `db:"provider" json:"provider"`
	Token      string    `db:"token" json:"token"`
	IsDefault  bool      `db:"is_default" json:"is_default"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
