package http

import (
	"context"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type AuthHandler struct {
	firebaseApp *firebase.App
	authService *application.AuthService
	cfg         *config.Configuration
}

func NewAuthHandler(app *firebase.App, authService *application.AuthService, cfg *config.Configuration) *AuthHandler {
	if app == nil || authService == nil || cfg == nil {
		return nil
	}
	return &AuthHandler{firebaseApp: app, authService: authService, cfg: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dtos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	authClient, err := h.firebaseApp.Auth(context.Background())
	if err != nil {
		response.ErrorResponse(c, "Failed to get Firebase Auth client", err, h.cfg.IsDevelopment())
		return
	}

	user, err := authClient.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.ErrorResponse(c, "Invalid email or password", err, h.cfg.IsDevelopment())
		return
	}

	// Generate access token
	accessToken, err := h.authService.GenerateAccessToken(c.Request.Context(), user.UID)
	if err != nil {
		response.ErrorResponse(c, "Failed to generate access token", err, h.cfg.IsDevelopment())
		return
	}

	// Get client information for refresh token
	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Create refresh token
	refreshToken, err := h.authService.CreateRefreshToken(c.Request.Context(), user.UID, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.ErrorResponse(c, "Failed to create refresh token", err, h.cfg.IsDevelopment())
		return
	}

	loginResponse := dtos.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.TokenHash, // Return the original token hash
		ExpiresIn:    3600,                   // 1 hour
		TokenType:    "Bearer",
		User: dtos.UserInfo{
			UID:         user.UID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			PhoneNumber: &user.PhoneNumber,
		},
	}

	response.SuccessResponseData(c, loginResponse)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	// Get Firebase Auth client
	authClient, err := h.firebaseApp.Auth(context.Background())
	if err != nil {
		response.ErrorResponse(c, "Failed to get Firebase Auth client", err, h.cfg.IsDevelopment())
		return
	}

	// Verify the refresh token to get user ID
	token, err := authClient.VerifyIDToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.ErrorResponse(c, "Invalid refresh token", err, h.cfg.IsDevelopment())
		return
	}

	// Validate the refresh token in our database
	refreshToken, err := h.authService.ValidateRefreshToken(c.Request.Context(), token.UID, req.RefreshToken)
	if err != nil {
		response.ErrorResponse(c, "Invalid or expired refresh token", err, h.cfg.IsDevelopment())
		return
	}

	// Generate new access token
	newAccessToken, err := h.authService.GenerateAccessToken(c.Request.Context(), token.UID)
	if err != nil {
		response.ErrorResponse(c, "Failed to generate new access token", err, h.cfg.IsDevelopment())
		return
	}

	// Get client information for new refresh token
	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Create new refresh token
	newRefreshToken, err := h.authService.CreateRefreshToken(c.Request.Context(), token.UID, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.ErrorResponse(c, "Failed to generate new refresh token", err, h.cfg.IsDevelopment())
		return
	}

	// Revoke the old refresh token
	if err := h.authService.RevokeRefreshToken(c.Request.Context(), refreshToken.ID); err != nil {
		// Log the error but don't fail the request
		// In production, you might want to log this
	}

	refreshResponse := dtos.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken.TokenHash,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
	}

	response.SuccessResponseData(c, refreshResponse)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dtos.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	// Get Firebase Auth client
	authClient, err := h.firebaseApp.Auth(context.Background())
	if err != nil {
		response.ErrorResponse(c, "Failed to get Firebase Auth client", err, h.cfg.IsDevelopment())
		return
	}

	// Verify the refresh token to get user ID
	token, err := authClient.VerifyIDToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.ErrorResponse(c, "Invalid refresh token", err, h.cfg.IsDevelopment())
		return
	}

	// Validate the refresh token in our database
	refreshToken, err := h.authService.ValidateRefreshToken(c.Request.Context(), token.UID, req.RefreshToken)
	if err != nil {
		response.ErrorResponse(c, "Invalid or expired refresh token", err, h.cfg.IsDevelopment())
		return
	}

	// Revoke the specific refresh token
	if err := h.authService.RevokeRefreshToken(c.Request.Context(), refreshToken.ID); err != nil {
		response.ErrorResponse(c, "Failed to revoke refresh token", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponse(c, "Successfully logged out", gin.H{
		"user_id": token.UID,
		"message": "Refresh token has been invalidated",
	})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("firebaseUID")
	if !exists {
		response.ErrorResponse(c, "User not authenticated", nil, h.cfg.IsDevelopment())
		return
	}

	authClient, err := h.firebaseApp.Auth(context.Background())
	if err != nil {
		response.ErrorResponse(c, "Failed to get Firebase Auth client", err, h.cfg.IsDevelopment())
		return
	}

	user, err := authClient.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		response.ErrorResponse(c, "Failed to get user information", err, h.cfg.IsDevelopment())
		return
	}

	userInfo := dtos.UserInfo{
		UID:         user.UID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		PhoneNumber: &user.PhoneNumber,
	}

	response.SuccessResponseData(c, userInfo)
}

// GetUserSessions returns all active sessions for the current user
func (h *AuthHandler) GetUserSessions(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("firebaseUID")
	if !exists {
		response.ErrorResponse(c, "User not authenticated", nil, h.cfg.IsDevelopment())
		return
	}

	// Get all user tokens
	tokens, err := h.authService.GetUserTokens(c.Request.Context(), userID.(string))
	if err != nil {
		response.ErrorResponse(c, "Failed to get user sessions", err, h.cfg.IsDevelopment())
		return
	}

	// Convert to session info
	var sessions []dtos.SessionInfo
	for _, token := range tokens {
		if !token.IsRevoked {
			session := dtos.SessionInfo{
				ID:         token.ID,
				DeviceInfo: token.DeviceInfo,
				IPAddress:  token.IPAddress,
				UserAgent:  token.UserAgent,
				CreatedAt:  token.CreatedAt.Format(time.RFC3339),
				ExpiresAt:  token.ExpiresAt.Format(time.RFC3339),
				IsCurrent:  false, // You could implement logic to determine current session
			}
			sessions = append(sessions, session)
		}
	}

	response.SuccessResponseData(c, sessions)
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	_, exists := c.Get("firebaseUID")
	if !exists {
		response.ErrorResponse(c, "User not authenticated", nil, h.cfg.IsDevelopment())
		return
	}

	var req dtos.RevokeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	// Revoke the session
	if err := h.authService.RevokeRefreshToken(c.Request.Context(), req.SessionID); err != nil {
		response.ErrorResponse(c, "Failed to revoke session", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponse(c, "Session revoked successfully", gin.H{
		"session_id": req.SessionID,
	})
}

// RevokeAllSessions revokes all sessions for the current user
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("firebaseUID")
	if !exists {
		response.ErrorResponse(c, "User not authenticated", nil, h.cfg.IsDevelopment())
		return
	}

	// Revoke all user tokens
	if err := h.authService.RevokeAllUserTokens(c.Request.Context(), userID.(string)); err != nil {
		response.ErrorResponse(c, "Failed to revoke all sessions", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponse(c, "All sessions revoked successfully", gin.H{
		"user_id": userID.(string),
	})
}
