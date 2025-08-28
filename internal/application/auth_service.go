package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/theHinneh/budgeting/internal/domain"
	"github.com/theHinneh/budgeting/internal/infrastructure/db/firebase"
)

type AuthService struct {
	refreshTokenRepo *firebase.RefreshTokenRepository
	userAuth         *firebase.FirebaseAuth
}

func NewAuthService(refreshTokenRepo *firebase.RefreshTokenRepository, userAuth *firebase.FirebaseAuth) *AuthService {
	return &AuthService{
		refreshTokenRepo: refreshTokenRepo,
		userAuth:         userAuth,
	}
}

// CreateRefreshToken creates a new refresh token for a user
func (s *AuthService) CreateRefreshToken(ctx context.Context, userID string, deviceInfo, ipAddress, userAgent string) (*domain.RefreshToken, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Set expiration (30 days)
	expiresAt := time.Now().AddDate(0, 0, 30)

	refreshToken := &domain.RefreshToken{
		UserID:     userID,
		TokenHash:  tokenString, // Will be hashed in repository
		IsRevoked:  false,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user ID
func (s *AuthService) ValidateRefreshToken(ctx context.Context, userID, tokenString string) (*domain.RefreshToken, error) {
	token, err := s.refreshTokenRepo.GetValidToken(ctx, userID, tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Check if token is revoked
	if token.IsRevoked {
		return nil, fmt.Errorf("refresh token revoked")
	}

	return token, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *AuthService) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	return s.refreshTokenRepo.RevokeToken(ctx, tokenID)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *AuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.refreshTokenRepo.RevokeAllUserTokens(ctx, userID)
}

// GetUserTokens returns all refresh tokens for a user
func (s *AuthService) GetUserTokens(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	return s.refreshTokenRepo.GetByUserID(ctx, userID)
}

// CleanupExpiredTokens removes expired tokens from the database
func (s *AuthService) CleanupExpiredTokens(ctx context.Context) error {
	return s.refreshTokenRepo.DeleteExpiredTokens(ctx)
}

// GenerateAccessToken generates a new Firebase custom token for access
func (s *AuthService) GenerateAccessToken(ctx context.Context, userID string) (string, error) {
	return s.userAuth.Auth.CustomToken(ctx, userID)
}
