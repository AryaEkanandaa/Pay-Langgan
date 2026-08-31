package models

import (
	"time"
)

type PaymentAttempt struct {
	ID            int       `db:"id" json:"id"`
	PaymentID     int       `db:"payment_id" json:"payment_id"`
	InvoiceID     int       `db:"invoice_id" json:"invoice_id"`
	Provider      string    `db:"provider" json:"provider"`
	AttemptNumber int       `db:"attempt_number" json:"attempt_number"`
	Status        string    `db:"status" json:"status"`
	ErrorMessage  *string   `db:"error_message" json:"error_message"`
	AttemptedAt   time.Time `db:"attempted_at" json:"attempted_at"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
