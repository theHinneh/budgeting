package domain

import "time"

type IncomeSource struct {
	UID       string
	UserID    string
	Source    string
	Amount    float64
	Currency  string
	Frequency string
	NextPayAt time.Time
	Active    bool
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
