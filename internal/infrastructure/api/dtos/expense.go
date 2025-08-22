package dtos

import (
	"time"

	"github.com/theHinneh/budgeting/internal/domain"
)

type AddExpenseRequest struct {
	Source   string  `json:"source" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency,omitempty"`
	Notes    string  `json:"notes,omitempty"`
}

type ExpenseResponse struct {
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListExpenseResponse struct {
	Expenses []*ExpenseResponse `json:"expenses"`
	Count    int                `json:"count"`
}

func NewExpenseResponse(expense *domain.Expense) *ExpenseResponse {
	if expense == nil {
		return nil
	}
	return &ExpenseResponse{
		UID:       expense.UID,
		UserID:    expense.UserID,
		Source:    expense.Source,
		Amount:    expense.Amount,
		Currency:  expense.Currency,
		Notes:     expense.Notes,
		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
	}
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
