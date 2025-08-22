package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
)

type ExpenseService struct {
	repo ports.ExpenseRepoPort
}

func NewExpenseService(repo ports.ExpenseRepoPort) *ExpenseService {
	return &ExpenseService{repo: repo}
}

var _ ports.ExpenseServicePort = (*ExpenseService)(nil)

func (s *ExpenseService) AddExpense(ctx context.Context, in ports.AddExpenseInput) (*domain.Expense, error) {
	userID := strings.TrimSpace(in.UserID)
	source := strings.TrimSpace(in.Source)
	currency := strings.TrimSpace(in.Currency)
	if userID == "" || source == "" || in.Amount <= 0 {
		return nil, ErrValidation
	}
	if currency == "" {
		currency = "USD"
	}

	expense := &domain.Expense{
		UID:       uuid.NewString(),
		UserID:    userID,
		Source:    source,
		Amount:    in.Amount,
		Currency:  currency,
		Notes:     strings.TrimSpace(in.Notes),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	return s.repo.CreateExpense(ctx, expense)
}

func (s *ExpenseService) ListExpenses(ctx context.Context, userID string) ([]*domain.Expense, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrValidation
	}
	return s.repo.ListExpensesByUser(ctx, userID)
}

func (s *ExpenseService) GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error) {
	userID = strings.TrimSpace(userID)
	expenseID = strings.TrimSpace(expenseID)
	if userID == "" || expenseID == "" {
		return nil, ErrValidation
	}
	return s.repo.GetExpense(ctx, userID, expenseID)
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, userID string, expenseID string, in ports.AddExpenseInput) (*domain.Expense, error) {
	userID = strings.TrimSpace(userID)
	expenseID = strings.TrimSpace(expenseID)
	source := strings.TrimSpace(in.Source)
	currency := strings.TrimSpace(in.Currency)

	if userID == "" || expenseID == "" || source == "" || in.Amount <= 0 {
		return nil, ErrValidation
	}
	if currency == "" {
		currency = "USD"
	}

	// Get existing expense
	expense, err := s.repo.GetExpense(ctx, userID, expenseID)
	if err != nil {
		return nil, err
	}

	// Update fields
	expense.Source = source
	expense.Amount = in.Amount
	expense.Currency = currency
	expense.Notes = strings.TrimSpace(in.Notes)
	expense.UpdatedAt = time.Now().UTC()

	return s.repo.UpdateExpense(ctx, expense)
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, userID string, expenseID string) error {
	userID = strings.TrimSpace(userID)
	expenseID = strings.TrimSpace(expenseID)
	if userID == "" || expenseID == "" {
		return ErrValidation
	}
	return s.repo.DeleteExpense(ctx, userID, expenseID)
}
