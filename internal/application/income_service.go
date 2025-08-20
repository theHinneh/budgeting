package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
)

type IncomeService struct {
	repo ports.IncomeRepoPort
}

func NewIncomeService(repo ports.IncomeRepoPort) *IncomeService {
	return &IncomeService{repo: repo}
}

var _ ports.IncomeServicePort = (*IncomeService)(nil)

func (s *IncomeService) AddIncome(ctx context.Context, in ports.AddIncomeInput) (*domain.Income, error) {
	userID := strings.TrimSpace(in.UserID)
	source := strings.TrimSpace(in.Source)
	currency := strings.TrimSpace(in.Currency)
	if userID == "" || source == "" || in.Amount <= 0 {
		return nil, ErrValidation
	}
	if currency == "" {
		currency = "USD"
	}

	income := &domain.Income{
		UID:       uuid.NewString(),
		UserID:    userID,
		Source:    source,
		Amount:    in.Amount,
		Currency:  currency,
		Notes:     strings.TrimSpace(in.Notes),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	return s.repo.CreateIncome(ctx, income)
}

func (s *IncomeService) ListIncomes(ctx context.Context, userID string) ([]*domain.Income, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrValidation
	}
	return s.repo.ListIncomesByUser(ctx, userID)
}

func (s *IncomeService) DeleteIncome(ctx context.Context, userID string, incomeID string) error {
	userID = strings.TrimSpace(userID)
	incomeID = strings.TrimSpace(incomeID)
	if userID == "" || incomeID == "" {
		return ErrValidation
	}
	// Load the income to identify its Source
	inc, err := s.repo.GetIncome(ctx, userID, incomeID)
	if err != nil {
		return err
	}
	// Delete the income entry
	if err := s.repo.DeleteIncome(ctx, userID, incomeID); err != nil {
		return err
	}
	// Also delete the corresponding income source(s) by Source string
	if inc != nil && strings.TrimSpace(inc.Source) != "" {
		_ = s.repo.DeleteIncomeSource(ctx, userID, inc.Source)
	}
	return nil
}

func (s *IncomeService) AddIncomeSource(ctx context.Context, in ports.AddIncomeSourceInput) (*domain.IncomeSource, error) {
	userID := strings.TrimSpace(in.UserID)
	source := strings.TrimSpace(in.Source)
	currency := strings.TrimSpace(in.Currency)
	freq := strings.ToLower(string(in.Frequency))
	if userID == "" || source == "" || in.Amount <= 0 || freq == "" {
		return nil, ErrValidation
	}
	if currency == "" {
		currency = "USD"
	}
	if !isValidFrequency(freq) {
		return nil, ErrValidation
	}
	next := time.Now().UTC()
	if in.NextPayAt != "" {
		parsedDate, _ := time.Parse("2006-01-02", in.NextPayAt)
		next = parsedDate.UTC()
	}
	src := &domain.IncomeSource{
		UID:       uuid.NewString(),
		UserID:    userID,
		Source:    source,
		Amount:    in.Amount,
		Currency:  currency,
		Frequency: freq,
		NextPayAt: next,
		Active:    true,
		Notes:     strings.TrimSpace(in.Notes),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	return s.repo.CreateIncomeSource(ctx, src)
}

func (s *IncomeService) ListIncomeSources(ctx context.Context, userID string) ([]*domain.IncomeSource, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrValidation
	}
	return s.repo.ListIncomeSourcesByUser(ctx, userID)
}

func (s *IncomeService) ProcessDueIncomes(ctx context.Context, userID string, now time.Time) (int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, ErrValidation
	}
	sources, err := s.repo.ListDueIncomeSources(ctx, userID, now)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, src := range sources {
		if src == nil || !src.Active {
			continue
		}
		next := src.NextPayAt.UTC()

		normalizedNext := time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
		normalizedNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		if normalizedNext.Equal(normalizedNow) {
			_, err := s.repo.CreateIncome(ctx, &domain.Income{
				UID:       uuid.NewString(),
				UserID:    userID,
				Source:    src.Source,
				Amount:    src.Amount,
				Currency:  src.Currency,
				Notes:     src.Notes,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			})
			if err != nil {
				return count, err
			}
			count++
			next = advanceNext(next, src.Frequency)
		}
		// Persist updated next pay date
		_ = s.repo.UpdateIncomeSource(ctx, userID, src.UID, map[string]interface{}{
			"NextPayAt": next,
			"UpdatedAt": time.Now().UTC(),
		})
	}
	return count, nil
}

func isValidFrequency(freq string) bool {
	switch freq {
	case string(ports.PayWeekly), string(ports.PayBiWeekly), string(ports.PayMonthly):
		return true
	default:
		return false
	}
}

func advanceNext(from time.Time, freq string) time.Time {
	switch freq {
	case string(ports.PayWeekly):
		return from.AddDate(0, 0, 7)
	case string(ports.PayBiWeekly):
		return from.AddDate(0, 0, 14)
	case string(ports.PayMonthly):
		return from.AddDate(0, 1, 0)
	default:
		return from.AddDate(0, 0, 7)
	}
}
