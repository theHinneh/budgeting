package domain

import "time"

type Expense struct {
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
