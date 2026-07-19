package models

import (
	"time"
)

type Product struct {
	ID          int       `db:"id" json:"id"`
	ServiceID   int       `db:"service_id" json:"service_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Status      string    `db:"status" json:"status"`
	Meta        JSONMap   `db:"meta" json:"meta"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
