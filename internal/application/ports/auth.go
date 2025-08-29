package ports

import (
	"context"

	"github.com/theHinneh/budgeting/internal/application/dto"
	"github.com/theHinneh/budgeting/internal/domain"
)

type AuthServicePort interface {
	// Authentication
	Login(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, deviceInfo, ipAddress, userAgent string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error

	// Session Management
	GetUserSessions(ctx context.Context, userID string) ([]*domain.RefreshToken, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	// Token Management
	ValidateRefreshToken(ctx context.Context, userID, tokenString string) (*domain.RefreshToken, error)
	CreateRefreshToken(ctx context.Context, userID string, deviceInfo, ipAddress, userAgent string) (*domain.RefreshToken, error)
	CleanupExpiredTokens(ctx context.Context) error
}
