package models

import (
	"time"
)

type Business struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Status    string    `db:"status" json:"status"`
	Deleted   bool      `db:"deleted" json:"deleted"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Meta      JSONMap   `db:"meta" json:"meta"`
}
