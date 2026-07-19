package models

import (
	"time"
)

type Subscription struct {
	ID              int        `db:"id" json:"id"`
	CustomerID      int        `db:"customer_id" json:"customer_id"`
	PlanID          int        `db:"plan_id" json:"plan_id"`
	Status          string     `db:"status" json:"status"`
	StartDate       time.Time  `db:"start_date" json:"start_date"`
	NextBillingDate *time.Time `db:"next_billing_date" json:"next_billing_date"`
	EndDate         *time.Time `db:"end_date" json:"end_date"`
	TrialEndsAt     *time.Time `db:"trial_ends_at" json:"trial_ends_at"`
	Meta            JSONMap    `db:"meta" json:"meta"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

type SubscriptionAddOn struct {
	ID             int       `db:"id" json:"id"`
	SubscriptionID int       `db:"subscription_id" json:"subscription_id"`
	AddOnID        int       `db:"add_on_id" json:"add_on_id"`
	Quantity       int       `db:"quantity" json:"quantity"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
