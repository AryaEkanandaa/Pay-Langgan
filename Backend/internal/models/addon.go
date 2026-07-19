package models

import (
	"time"
)

type AddOn struct {
	ID           int       `db:"id" json:"id"`
	ProductID    int       `db:"product_id" json:"product_id"`
	Name         string    `db:"name" json:"name"`
	Price        float64   `db:"price" json:"price"`
	BillingCycle string    `db:"billing_cycle" json:"billing_cycle"`
	Meta         JSONMap   `db:"meta" json:"meta"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
