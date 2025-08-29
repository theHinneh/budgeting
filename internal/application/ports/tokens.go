package ports

import (
	"context"

	"github.com/theHinneh/budgeting/internal/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByID(ctx context.Context, id string) (*domain.RefreshToken, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.RefreshToken, error)
	GetValidToken(ctx context.Context, userID, tokenHash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, id string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteToken(ctx context.Context, id string) error
}

type TokenAuthenticator interface {
	CreateCustomToken(ctx context.Context, userID string) (string, error)
	VerifyIDToken(ctx context.Context, idToken string) (string, error) // Returns userID
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type TokenGenerator interface {
	GenerateSecureToken() (string, error)
	HashToken(token string) string
}
