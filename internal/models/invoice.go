package models

import (
	"time"
)

type Invoice struct {
	ID              int            `db:"id" json:"id"`
	SubscriptionID  int            `db:"subscription_id" json:"subscription_id"`
	InvoiceNumber   string         `db:"invoice_number" json:"invoice_number"`
	Amount          float64        `db:"amount" json:"amount"`
	Status          string         `db:"status" json:"status"`
	DueDate         *time.Time     `db:"due_date" json:"due_date"`
	PaidAt          *time.Time     `db:"paid_at" json:"paid_at"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
}
