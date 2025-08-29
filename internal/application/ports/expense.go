package ports

import (
	"context"
	"time"

	"github.com/theHinneh/budgeting/internal/application/dto"
	"github.com/theHinneh/budgeting/internal/domain"
)

type ExpenseServicePort interface {
	AddExpense(ctx context.Context, in dto.AddExpenseInput) (*domain.Expense, error)
	ListExpenses(ctx context.Context, userID string) ([]*domain.Expense, error)
	GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error)
	UpdateExpense(ctx context.Context, userID string, expenseID string, in dto.AddExpenseInput) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, userID string, expenseID string) error
	ProcessDueExpenses(ctx context.Context, userID string, now time.Time) (int, error)
}

type ExpenseRepoPort interface {
	CreateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error)
	ListExpensesByUser(ctx context.Context, userID string) ([]*domain.Expense, error)
	GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error)
	UpdateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, userID string, expenseID string) error
	ListRecurringExpenses(ctx context.Context, userID string, before time.Time) ([]*domain.Expense, error)
	UpdateExpenseRecurringStatus(ctx context.Context, userID string, expenseID string, nextOccurrenceDate time.Time) error
}
