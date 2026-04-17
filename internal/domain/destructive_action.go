package domain

import "time"

type DestructiveAction struct {
	ID                     int16     `db:"id"                       json:"id"`
	Code                   string    `db:"code"                     json:"code"`
	Name                   string    `db:"name"                     json:"name"`
	AffectsConfidentiality bool      `db:"affects_confidentiality"  json:"affects_confidentiality"`
	AffectsIntegrity       bool      `db:"affects_integrity"        json:"affects_integrity"`
	AffectsAvailability    bool      `db:"affects_availability"     json:"affects_availability"`
	Description            string    `db:"description"              json:"description,omitempty"`
	CreatedAt              time.Time `db:"created_at"               json:"created_at"`
}
