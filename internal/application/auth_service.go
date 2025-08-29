package application

import (
	"context"
	"fmt"
	"time"

	"github.com/theHinneh/budgeting/internal/application/dto"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
	"github.com/theHinneh/budgeting/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type AuthService struct {
	refreshTokenRepo ports.RefreshTokenRepository
	tokenAuth        ports.TokenAuthenticator
	tokenGenerator   ports.TokenGenerator
}

func NewAuthService(
	refreshTokenRepo ports.RefreshTokenRepository,
	tokenAuth ports.TokenAuthenticator,
	tokenGenerator ports.TokenGenerator,
) ports.AuthServicePort {
	return &AuthService{
		refreshTokenRepo: refreshTokenRepo,
		tokenAuth:        tokenAuth,
		tokenGenerator:   tokenGenerator,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	user, err := s.tokenAuth.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password: %w", err)
	}

	accessToken, err := s.tokenAuth.CreateCustomToken(ctx, user.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.CreateRefreshToken(ctx, user.UID, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.TokenHash,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
		User: &dto.UserInfo{
			UID:         user.UID,
			Email:       user.Email,
			DisplayName: user.FirstName + " " + user.LastName,
			PhoneNumber: user.PhoneNumber,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, deviceInfo, ipAddress, userAgent string) (*dto.RefreshTokenResponse, error) {
	userID, err := s.tokenAuth.VerifyIDToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	oldToken, err := s.ValidateRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	newAccessToken, err := s.tokenAuth.CreateCustomToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := s.CreateRefreshToken(ctx, userID, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	if err := s.RevokeSession(ctx, oldToken.ID); err != nil {
		logger.Error("failed to revoke old refresh token", zap.Error(err))
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken.TokenHash,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {

	userID, err := s.tokenAuth.VerifyIDToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	token, err := s.ValidateRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	return s.RevokeSession(ctx, token.ID)
}

func (s *AuthService) GetUserSessions(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	return s.refreshTokenRepo.GetByUserID(ctx, userID)
}

func (s *AuthService) RevokeSession(ctx context.Context, sessionID string) error {
	return s.refreshTokenRepo.RevokeToken(ctx, sessionID)
}

func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return s.refreshTokenRepo.RevokeAllUserTokens(ctx, userID)
}

func (s *AuthService) CreateRefreshToken(ctx context.Context, userID string, deviceInfo, ipAddress, userAgent string) (*domain.RefreshToken, error) {

	tokenString, err := s.tokenGenerator.GenerateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().AddDate(0, 0, 30)

	refreshToken := &domain.RefreshToken{
		UserID:     userID,
		TokenHash:  tokenString,
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

func (s *AuthService) ValidateRefreshToken(ctx context.Context, userID, tokenString string) (*domain.RefreshToken, error) {
	token, err := s.refreshTokenRepo.GetValidToken(ctx, userID, tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	if token.IsRevoked {
		return nil, fmt.Errorf("refresh token revoked")
	}

	return token, nil
}

func (s *AuthService) CleanupExpiredTokens(ctx context.Context) error {
	return s.refreshTokenRepo.DeleteExpiredTokens(ctx)
}
