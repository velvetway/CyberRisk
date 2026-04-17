package domain

import "time"

type ThreatSource struct {
	ID          int16     `db:"id"          json:"id"`
	Code        string    `db:"code"        json:"code"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
}
