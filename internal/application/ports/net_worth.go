package ports

import (
	"context"

	"github.com/theHinneh/budgeting/internal/application/dto"
)

type NetWorthServicePort interface {
	GetNetWorth(ctx context.Context, userID string) (*dto.NetWorthResponse, error)
}
