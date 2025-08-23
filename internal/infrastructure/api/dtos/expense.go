package dtos

import (
	"time"

	"github.com/theHinneh/budgeting/internal/domain"
)

type AddExpenseRequest struct {
	Source              string  `json:"source" binding:"required"`
	Amount              float64 `json:"amount" binding:"required,gt=0"`
	Currency            string  `json:"currency,omitempty"`
	Notes               string  `json:"notes,omitempty"`
	IsRecurring         bool    `json:"is_recurring,omitempty"`
	RecurrenceFrequency string  `json:"recurrence_frequency,omitempty"`
	NextOccurrenceDate  string  `json:"next_occurrence_date,omitempty"`
}

func (r *AddExpenseRequest) ToDomain() *domain.Expense {
	next := time.Now().UTC()
	if r.NextOccurrenceDate != "" {
		parsedDate, _ := time.Parse("2006-01-02", r.NextOccurrenceDate)
		next = parsedDate.UTC()
	}

	expense := &domain.Expense{
		Source:              r.Source,
		Amount:              r.Amount,
		Currency:            r.Currency,
		Notes:               r.Notes,
		IsRecurring:         r.IsRecurring,
		RecurrenceFrequency: r.RecurrenceFrequency,
		NextOccurrenceDate:  next,
	}

	return expense
}

type ExpenseResponse struct {
	UID                 string    `json:"uid"`
	UserID              string    `json:"user_id"`
	Source              string    `json:"source"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency,omitempty"`
	Notes               string    `json:"notes,omitempty"`
	IsRecurring         bool      `json:"is_recurring"`
	RecurrenceFrequency string    `json:"recurrence_frequency,omitempty"`
	NextOccurrenceDate  time.Time `json:"next_occurrence_date,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewExpenseResponse(expense *domain.Expense) *ExpenseResponse {
	if expense == nil {
		return nil
	}
	return &ExpenseResponse{
		UID:                 expense.UID,
		UserID:              expense.UserID,
		Source:              expense.Source,
		Amount:              expense.Amount,
		Currency:            expense.Currency,
		Notes:               expense.Notes,
		IsRecurring:         expense.IsRecurring,
		RecurrenceFrequency: expense.RecurrenceFrequency,
		NextOccurrenceDate:  expense.NextOccurrenceDate,
		CreatedAt:           expense.CreatedAt,
		UpdatedAt:           expense.UpdatedAt,
	}
}

type ListExpenseResponse struct {
	Expenses []*ExpenseResponse `json:"expenses"`
	Count    int                `json:"count"`
}

func NewListExpenseResponse(expenses []*domain.Expense) *ListExpenseResponse {
	resps := make([]*ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		resps[i] = NewExpenseResponse(expense)
	}
	return &ListExpenseResponse{
		Expenses: resps,
		Count:    len(resps),
	}
}
