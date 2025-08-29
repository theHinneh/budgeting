package http

import (
	"context"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type AuthHandler struct {
	firebaseApp *firebase.App
	authService ports.AuthServicePort
	cfg         *config.Configuration
}

func NewAuthHandler(app *firebase.App, authService ports.AuthServicePort, cfg *config.Configuration) *AuthHandler {
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

	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	loginResponse, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.ErrorResponse(c, "Login failed", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponseData(c, loginResponse)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	refreshResponse, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.ErrorResponse(c, "Token refresh failed", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponseData(c, refreshResponse)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dtos.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	authClient, err := h.firebaseApp.Auth(context.Background())
	if err != nil {
		response.ErrorResponse(c, "Failed to get Firebase Auth client", err, h.cfg.IsDevelopment())
		return
	}

	token, err := authClient.VerifyIDToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.ErrorResponse(c, "Invalid refresh token", err, h.cfg.IsDevelopment())
		return
	}

	refreshToken, err := h.authService.ValidateRefreshToken(c.Request.Context(), token.UID, req.RefreshToken)
	if err != nil {
		response.ErrorResponse(c, "Invalid or expired refresh token", err, h.cfg.IsDevelopment())
		return
	}

	if err := h.authService.RevokeSession(c.Request.Context(), refreshToken.ID); err != nil {
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

func (h *AuthHandler) GetUserSessions(c *gin.Context) {
	userID, exists := c.Get("firebaseUID")
	if !exists {
		response.ErrorResponse(c, "User not authenticated", nil, h.cfg.IsDevelopment())
		return
	}

	tokens, err := h.authService.GetUserSessions(c.Request.Context(), userID.(string))
	if err != nil {
		response.ErrorResponse(c, "Failed to get user sessions", err, h.cfg.IsDevelopment())
		return
	}

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
				IsCurrent:  false,
			}
			sessions = append(sessions, session)
		}
	}

	response.SuccessResponseData(c, sessions)
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
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

	if err := h.authService.RevokeSession(c.Request.Context(), req.SessionID); err != nil {
		response.ErrorResponse(c, "Failed to revoke session", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponse(c, "Session revoked successfully", gin.H{
		"session_id": req.SessionID,
	})
}

func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userID, exists := c.Get("firebaseUID")
	if !exists {
		response.ErrorResponse(c, "User not authenticated", nil, h.cfg.IsDevelopment())
		return
	}

	if err := h.authService.RevokeAllUserSessions(c.Request.Context(), userID.(string)); err != nil {
		response.ErrorResponse(c, "Failed to revoke all sessions", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponse(c, "All sessions revoked successfully", gin.H{
		"user_id": userID.(string),
	})
}
