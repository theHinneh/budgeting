package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/core/ports"
	"github.com/theHinneh/budgeting/pkg/logger"
	"github.com/theHinneh/budgeting/pkg/response"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	Service ports.UserServicePort
}

func NewUserHandler(svc ports.UserServicePort) *UserHandler {
	if svc == nil {
		return nil
	}
	return &UserHandler{Service: svc}
}

// RegisterUserRoutes registers user CRUD HTTP routes.
func RegisterUserRoutes(router *gin.Engine, uh *UserHandler) {
	logger.Info("Registering user routes...")
	if router == nil || uh == nil {
		return
	}
	g := router.Group("/users")
	{
		g.POST("", uh.CreateUser)
		g.GET("/:id", uh.GetUser)
		g.PUT("/:id", uh.UpdateUser)
		g.DELETE("/:id", uh.DeleteUser)
	}
}

type createUserRequest struct {
	UID         string  `json:"uid"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	FirstName   string  `json:"firstname"`
	LastName    string  `json:"lastname"`
	PhoneNumber *string `json:"phone_number"`
	Password    string  `json:"password,omitempty"`
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
		response.ErrorResponse(c, "invalid request body", err)
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		response.ErrorResponse(c, "username, email, firstname, and lastname are required", nil)
		return
	}

	ctx := c.Request.Context()
	uid, err := h.Service.CreateUser(ctx, ports.CreateUserInput{
		UID:         strings.TrimSpace(req.UID),
		Username:    strings.TrimSpace(req.Username),
		Email:       strings.TrimSpace(req.Email),
		FirstName:   strings.TrimSpace(req.FirstName),
		LastName:    strings.TrimSpace(req.LastName),
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
	})
	if err != nil {
		logger.Error("create user failed", zap.Error(err))
		response.ErrorResponse(c, "failed to create user", err)
		return
	}

	response.SuccessWithStatusResponse(c, http.StatusCreated, "user created", gin.H{"uid": uid})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	uid := strings.TrimSpace(c.Param("id"))
	if uid == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}

	ctx := c.Request.Context()
	user, err := h.Service.GetUser(ctx, uid)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			response.NotFoundResponse(c, "user not found", err)
			return
		}
		response.ErrorResponse(c, "failed to get user", err)
		return
	}
	response.SuccessResponseData(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	uid := strings.TrimSpace(c.Param("id"))
	if uid == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err)
		return
	}

	ctx := c.Request.Context()
	user, err := h.Service.UpdateUser(ctx, uid, ports.UpdateUserInput{
		Username:    req.Username,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		logger.Error("failed to update user", zap.String("uid", uid), zap.Error(err))
		response.ErrorResponse(c, "failed to update user", err)
		return
	}
	response.SuccessResponseData(c, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	uid := strings.TrimSpace(c.Param("id"))
	if uid == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}
	ctx := c.Request.Context()
	if err := h.Service.DeleteUser(ctx, uid); err != nil {
		logger.Error("failed to delete user", zap.String("uid", uid), zap.Error(err))
		response.ErrorResponse(c, "failed to delete user", err)
		return
	}
	response.SuccessResponse(c, "user deleted", gin.H{"uid": uid})
}
