package models

import (
	"time"
)

type Service struct {
	ID          int       `db:"id" json:"id"`
	BusinessID  string    `db:"business_id" json:"business_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Meta        JSONMap   `db:"meta" json:"meta"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
