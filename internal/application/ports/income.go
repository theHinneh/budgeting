package ports

import (
	"context"
	"time"

	"github.com/theHinneh/budgeting/internal/application/dto"
	"github.com/theHinneh/budgeting/internal/domain"
)

type IncomeServicePort interface {
	AddIncome(ctx context.Context, in dto.AddIncomeInput) (*domain.Income, error)
	ListIncomes(ctx context.Context, userID string) ([]*domain.Income, error)
	DeleteIncome(ctx context.Context, userID string, incomeID string) error

	AddIncomeSource(ctx context.Context, in dto.AddIncomeSourceInput) (*domain.IncomeSource, error)
	ListIncomeSources(ctx context.Context, userID string) ([]*domain.IncomeSource, error)
	ProcessDueIncomes(ctx context.Context, userID string, now time.Time) (int, error)
}

type IncomeRepoPort interface {
	CreateIncome(ctx context.Context, income *domain.Income) (*domain.Income, error)
	ListIncomesByUser(ctx context.Context, userID string) ([]*domain.Income, error)
	GetIncome(ctx context.Context, userID string, incomeID string) (*domain.Income, error)
	DeleteIncome(ctx context.Context, userID string, incomeID string) error

	CreateIncomeSource(ctx context.Context, src *domain.IncomeSource) (*domain.IncomeSource, error)
	ListIncomeSourcesByUser(ctx context.Context, userID string) ([]*domain.IncomeSource, error)
	ListDueIncomeSources(ctx context.Context, userID string, before time.Time) ([]*domain.IncomeSource, error)
	UpdateIncomeSource(ctx context.Context, userID string, id string, updates map[string]interface{}) error
	DeleteIncomeSource(ctx context.Context, userID string, source string) error
}
