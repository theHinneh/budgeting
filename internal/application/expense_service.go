package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theHinneh/budgeting/internal/application/dto"
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

func (s *ExpenseService) AddExpense(ctx context.Context, in dto.AddExpenseInput) (*domain.Expense, error) {
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
		UID:                 uuid.NewString(),
		UserID:              userID,
		Source:              source,
		Amount:              in.Amount,
		Currency:            currency,
		Notes:               strings.TrimSpace(in.Notes),
		IsRecurring:         in.IsRecurring,
		RecurrenceFrequency: strings.TrimSpace(in.RecurrenceFrequency),
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if in.NextOccurrenceDate != nil {
		expense.NextOccurrenceDate = *in.NextOccurrenceDate
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

func (s *ExpenseService) UpdateExpense(ctx context.Context, userID string, expenseID string, in dto.AddExpenseInput) (*domain.Expense, error) {
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

	expense, err := s.repo.GetExpense(ctx, userID, expenseID)
	if err != nil {
		return nil, err
	}

	expense.Source = source
	expense.Amount = in.Amount
	expense.Currency = currency
	expense.Notes = strings.TrimSpace(in.Notes)
	expense.IsRecurring = in.IsRecurring
	expense.RecurrenceFrequency = strings.TrimSpace(in.RecurrenceFrequency)

	if in.NextOccurrenceDate != nil {
		expense.NextOccurrenceDate = *in.NextOccurrenceDate
	}

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

func (s *ExpenseService) ProcessDueExpenses(ctx context.Context, userID string, now time.Time) (int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, ErrValidation
	}

	expenses, err := s.repo.ListRecurringExpenses(ctx, userID, now)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, exp := range expenses {
		if exp == nil || !exp.IsRecurring {
			continue
		}

		next := exp.NextOccurrenceDate.UTC()

		normalizedNext := time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
		normalizedNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		if normalizedNext.Equal(normalizedNow) {
			_, err := s.repo.CreateExpense(ctx, &domain.Expense{
				UID:       uuid.NewString(),
				UserID:    userID,
				Source:    exp.Source,
				Amount:    exp.Amount,
				Currency:  exp.Currency,
				Notes:     exp.Notes,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			})
			if err != nil {
				return count, err
			}
			count++
			next = advanceExpenseNextOccurrence(next, exp.RecurrenceFrequency)
		}

		_ = s.repo.UpdateExpenseRecurringStatus(ctx, userID, exp.UID, next)
	}

	return count, nil
}

func isValidExpenseFrequency(freq string) bool {
	switch freq {
	case string(dto.RecurringWeekly), string(dto.RecurringBiWeekly), string(dto.RecurringMonthly), string(dto.RecurringAnnually):
		return true
	default:
		return false
	}
}

func advanceExpenseNextOccurrence(from time.Time, freq string) time.Time {
	switch freq {
	case string(dto.RecurringWeekly):
		return from.AddDate(0, 0, 7)
	case string(dto.RecurringBiWeekly):
		return from.AddDate(0, 0, 14)
	case string(dto.RecurringMonthly):
		return from.AddDate(0, 1, 0)
	case string(dto.RecurringAnnually):
		return from.AddDate(1, 0, 0)
	default:
		return from.AddDate(0, 0, 7)
	}
}
