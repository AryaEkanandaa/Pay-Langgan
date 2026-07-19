package models

import (
	"time"
)

type Customer struct {
	ID         int       `db:"id" json:"id"`
	BusinessID string    `db:"business_id" json:"business_id"`
	Name       string    `db:"name" json:"name"`
	Email      *string   `db:"email" json:"email"`
	Contact    *string   `db:"contact" json:"contact"`
	Meta       JSONMap   `db:"meta" json:"meta"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
