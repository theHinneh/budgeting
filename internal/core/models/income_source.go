package models

import (
	"time"
)

type IncomeSource struct {
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency,omitempty"`
	Frequency string    `json:"frequency"`   // e.g., "weekly", "biweekly", "monthly"
	NextPayAt time.Time `json:"next_pay_at"` // when the next payout is due
	Active    bool      `json:"active"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
