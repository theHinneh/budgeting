package dto

import "time"

type AddExpenseInput struct {
	UserID              string
	Source              string
	Amount              float64
	Currency            string
	Notes               string
	IsRecurring         bool
	RecurrenceFrequency string
	NextOccurrenceDate  *time.Time
}

type ExpenseRecurrenceFrequency string

const (
	RecurringWeekly   ExpenseRecurrenceFrequency = "weekly"
	RecurringBiWeekly ExpenseRecurrenceFrequency = "biweekly"
	RecurringMonthly  ExpenseRecurrenceFrequency = "monthly"
	RecurringAnnually ExpenseRecurrenceFrequency = "annually"
)
