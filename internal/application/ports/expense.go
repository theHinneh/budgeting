package ports

import (
	"context"

	"github.com/theHinneh/budgeting/internal/domain"
)

type AddExpenseInput struct {
	UserID   string
	Source   string
	Amount   float64
	Currency string
	Notes    string
}

type ExpenseServicePort interface {
	AddExpense(ctx context.Context, in AddExpenseInput) (*domain.Expense, error)
	ListExpenses(ctx context.Context, userID string) ([]*domain.Expense, error)
	GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error)
	UpdateExpense(ctx context.Context, userID string, expenseID string, in AddExpenseInput) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, userID string, expenseID string) error
}

type ExpenseRepoPort interface {
	CreateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error)
	ListExpensesByUser(ctx context.Context, userID string) ([]*domain.Expense, error)
	GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error)
	UpdateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, userID string, expenseID string) error
}
