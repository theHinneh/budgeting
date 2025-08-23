package domain

import "time"

type Expense struct {
	UID                 string
	UserID              string
	Source              string
	Amount              float64
	Currency            string
	Notes               string
	IsRecurring         bool
	RecurrenceFrequency string
	NextOccurrenceDate  time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
