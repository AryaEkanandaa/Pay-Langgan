package models

import "time"

type InvoiceListItem struct {
	ID             int        `db:"id" json:"id"`
	SubscriptionID int        `db:"subscription_id" json:"subscription_id"`
	InvoiceNumber  string     `db:"invoice_number" json:"invoice_number"`
	Amount         float64    `db:"amount" json:"amount"`
	Status         string     `db:"status" json:"status"`
	DueDate        *time.Time `db:"due_date" json:"due_date"`
	PaidAt         *time.Time `db:"paid_at" json:"paid_at"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	CustomerID     int        `db:"customer_id" json:"customer_id"`
	CustomerName   string     `db:"customer_name" json:"customer_name"`
	CustomerEmail  *string    `db:"customer_email" json:"customer_email"`
	PlanName       string     `db:"plan_name" json:"plan_name"`
}

type InvoiceItemResponse struct {
	ID          int       `db:"id" json:"id"`
	InvoiceID   int       `db:"invoice_id" json:"invoice_id"`
	ItemType    string    `db:"item_type" json:"item_type"`
	Description string    `db:"description" json:"description"`
	Quantity    int       `db:"quantity" json:"quantity"`
	UnitPrice   float64   `db:"unit_price" json:"unit_price"`
	Subtotal    float64   `db:"subtotal" json:"subtotal"`
	Meta        JSONMap   `db:"meta" json:"meta"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type InvoiceDetailResponse struct {
	InvoiceListItem
	Items   []InvoiceItemResponse `json:"items"`
	Payment *Payment              `json:"payment,omitempty"`
}
