package http

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	Service     ports.UserServicePort
	firebaseApp *firebase.App
	cfg         *config.Configuration
}

func NewUserHandler(svc ports.UserServicePort, app *firebase.App, cfg *config.Configuration) *UserHandler {
	if svc == nil || app == nil || cfg == nil {
		return nil
	}
	return &UserHandler{Service: svc, firebaseApp: app, cfg: cfg}
}

type createUserRequest struct {
	Username    string  `json:"username" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	FirstName   string  `json:"firstname" binding:"required"`
	LastName    string  `json:"lastname" binding:"required"`
	PhoneNumber *string `json:"phone_number"`
	Password    string  `json:"password,omitempty" binding:"required,min=6"`
}

type updateUserRequest struct {
	Username    *string `json:"username"`
	Email       *string `json:"email"`
	FirstName   *string `json:"firstname"`
	LastName    *string `json:"lastname"`
	PhoneNumber *string `json:"phone_number"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	// Create Firebase Auth user
	authClient, err := h.firebaseApp.Auth(context.Background())
	if err != nil {
		response.ErrorResponse(c, "Failed to get Firebase Auth client", err, h.cfg.IsDevelopment())
		return
	}

	displayName := strings.TrimSpace(req.FirstName + " " + req.LastName)
	params := (&fbAuth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		DisplayName(displayName)
	if req.PhoneNumber != nil && strings.TrimSpace(*req.PhoneNumber) != "" {
		params = params.PhoneNumber(*req.PhoneNumber)
	}

	fbUser, err := authClient.CreateUser(c.Request.Context(), params)
	if err != nil {
		response.ErrorResponse(c, "Failed to create Firebase Auth user", err, h.cfg.IsDevelopment())
		return
	}

	// Create user profile in our database using the Firebase UID
	uid, err := h.Service.CreateUser(c.Request.Context(), ports.CreateUserInput{
		UID:         fbUser.UID,
		Username:    strings.TrimSpace(req.Username),
		Email:       strings.TrimSpace(req.Email),
		FirstName:   strings.TrimSpace(req.FirstName),
		LastName:    strings.TrimSpace(req.LastName),
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		// If user creation in our database fails, attempt to delete Firebase Auth user
		_ = authClient.DeleteUser(context.Background(), fbUser.UID)
		response.ErrorResponse(c, "failed to create user profile", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessWithStatusResponse(c, http.StatusCreated, "user created", gin.H{"uid": uid})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	requestedUID := strings.TrimSpace(c.Param("id"))
	if requestedUID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to user data", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx := c.Request.Context()
	user, err := h.Service.GetUser(ctx, requestedUID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			response.NotFoundResponse(c, "user not found", err, h.cfg.IsDevelopment())
			return
		}
		response.ErrorResponse(c, "failed to get user", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponseData(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	requestedUID := strings.TrimSpace(c.Param("id"))
	if requestedUID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to user data", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	ctx := c.Request.Context()
	user, err := h.Service.UpdateUser(ctx, requestedUID, ports.UpdateUserInput{
		Username:    req.Username,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		response.ErrorResponse(c, "failed to update user", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponseData(c, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	requestedUID := strings.TrimSpace(c.Param("id"))
	if requestedUID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to user data", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx := c.Request.Context()
	if err := h.Service.DeleteUser(ctx, requestedUID); err != nil {
		response.ErrorResponse(c, "failed to delete user", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponse(c, "user deleted", gin.H{"uid": requestedUID})
}

func (h *UserHandler) ForgotPassword(c *gin.Context) {
	type forgotPasswordRequest struct {
		Email string `json:"email"`
	}
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		response.ErrorResponse(c, "email is required", nil, h.cfg.IsDevelopment())
		return
	}
	link, err := h.Service.ForgotPassword(c.Request.Context(), email)
	if err != nil {
		response.ErrorResponse(c, "failed to generate reset link", err, h.cfg.IsDevelopment())
		return
	}
	// showing link for easy test in Postman
	response.SuccessResponse(c, "password reset link generated", gin.H{"reset_link": link})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	requestedUID := strings.TrimSpace(c.Param("id"))
	if requestedUID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to change password", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	type changePasswordRequest struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}
	newPwd := req.NewPassword

	if err := h.Service.ChangePassword(c.Request.Context(), requestedUID, newPwd); err != nil {
		response.ErrorResponse(c, "failed to change password", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponse(c, "password changed", gin.H{"uid": requestedUID})
}
