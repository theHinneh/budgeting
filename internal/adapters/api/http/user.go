package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/adapters/db/firebase"
	"github.com/theHinneh/budgeting/internal/core/models"
	"github.com/theHinneh/budgeting/pkg/logger"
	"github.com/theHinneh/budgeting/pkg/response"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHandler provides CRUD endpoints for users backed by Firebase Auth + Firestore.
type UserHandler struct {
	Auth      *fbAuth.Client
	Firestore *firestore.Client
}

func NewUserHandler(fb *firebase.Database) *UserHandler {
	if fb == nil {
		return nil
	}
	return &UserHandler{Auth: fb.Auth, Firestore: fb.Firestore}
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

// createUserRequest models the expected JSON for user creation.
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
	var uid string

	// If password is provided, create the user in Firebase Auth.
	if strings.TrimSpace(req.Password) != "" {
		params := (&fbAuth.UserToCreate{}).
			Email(req.Email).
			Password(req.Password).
			DisplayName(strings.TrimSpace(req.FirstName + " " + req.LastName))
		if req.PhoneNumber != nil && strings.TrimSpace(*req.PhoneNumber) != "" {
			params = params.PhoneNumber(*req.PhoneNumber)
		}
		u, err := h.Auth.CreateUser(ctx, params)
		if err != nil {
			logger.Error("firebase auth create user failed", zap.Error(err))
			response.ErrorResponse(c, "failed to create auth user", err)
			return
		}
		uid = u.UID
	} else {
		// No password: expect that the Auth user is already created (e.g., Google Sign-In).
		uid = strings.TrimSpace(req.UID)
		if uid == "" {
			response.ErrorResponse(c, "uid is required when password is not provided", nil)
			return
		}
		// Optional: verify the auth user exists
		if _, err := h.Auth.GetUser(ctx, uid); err != nil {
			logger.Error("auth user not found for provided uid", zap.String("uid", uid), zap.Error(err))
			response.ErrorResponse(c, "auth user not found for provided uid", err)
			return
		}
	}

	user := models.NewUser(uid, req.Username, req.Email, req.FirstName, req.LastName, req.PhoneNumber)

	// Store profile in Firestore
	if err := h.saveProfile(ctx, user); err != nil {
		response.ErrorResponse(c, "failed to save user profile", err)
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
	user, err := h.getProfile(ctx, uid)
	if err != nil {
		// Firestore returns gRPC NotFound error
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
	var updates []firestore.Update

	if req.Username != nil {
		updates = append(updates, firestore.Update{Path: "Username", Value: *req.Username})
	}
	if req.Email != nil {
		updates = append(updates, firestore.Update{Path: "Email", Value: *req.Email})
	}
	if req.FirstName != nil {
		updates = append(updates, firestore.Update{Path: "FirstName", Value: *req.FirstName})
	}
	if req.LastName != nil {
		updates = append(updates, firestore.Update{Path: "LastName", Value: *req.LastName})
	}
	if req.PhoneNumber != nil {
		updates = append(updates, firestore.Update{Path: "PhoneNumber", Value: req.PhoneNumber})
	}
	updates = append(updates, firestore.Update{Path: "UpdatedAt", Value: time.Now().UTC()})

	// Update Firestore profile if there are any changes beyond UpdatedAt
	if len(updates) > 1 {
		if _, err := h.Firestore.Collection("users").Doc(uid).Update(ctx, updates); err != nil {
			response.ErrorResponse(c, "failed to update user profile", err)
			return
		}
	}

	// Update Auth fields if relevant
	var authUpdate *fbAuth.UserToUpdate
	if req.Email != nil {
		if authUpdate == nil {
			authUpdate = &fbAuth.UserToUpdate{}
		}
		authUpdate = authUpdate.Email(*req.Email)
	}
	if req.FirstName != nil || req.LastName != nil {
		fn := ""
		ln := ""
		if req.FirstName != nil {
			fn = *req.FirstName
		}
		if req.LastName != nil {
			ln = *req.LastName
		}
		dn := strings.TrimSpace(strings.TrimSpace(fn + " " + ln))
		if dn != "" {
			if authUpdate == nil {
				authUpdate = &fbAuth.UserToUpdate{}
			}
			authUpdate = authUpdate.DisplayName(dn)
		}
	}
	if req.PhoneNumber != nil {
		if authUpdate == nil {
			authUpdate = &fbAuth.UserToUpdate{}
		}
		if strings.TrimSpace(*req.PhoneNumber) == "" {
			// Clearing phone number is not directly supported with empty string; skip
		} else {
			authUpdate = authUpdate.PhoneNumber(*req.PhoneNumber)
		}
	}
	if authUpdate != nil {
		if _, err := h.Auth.UpdateUser(ctx, uid, authUpdate); err != nil {
			logger.Error("failed to update auth user", zap.String("uid", uid), zap.Error(err))
			response.ErrorResponse(c, "failed to update auth user", err)
			return
		}
	}

	user, err := h.getProfile(ctx, uid)
	if err != nil {
		response.ErrorResponse(c, "failed to fetch updated user", err)
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

	// Delete profile doc first
	if _, err := h.Firestore.Collection("users").Doc(uid).Delete(ctx); err != nil {
		// If not found, continue to attempt to delete auth user
		logger.Debug("failed to delete user profile doc", zap.String("uid", uid), zap.Error(err))
	}
	// Delete auth user
	if err := h.Auth.DeleteUser(ctx, uid); err != nil {
		logger.Error("failed to delete auth user", zap.String("uid", uid), zap.Error(err))
		response.ErrorResponse(c, "failed to delete auth user", err)
		return
	}

	response.SuccessResponse(c, "user deleted", gin.H{"uid": uid})
}

func (h *UserHandler) saveProfile(ctx context.Context, u *models.User) error {
	if u == nil || strings.TrimSpace(u.UID) == "" {
		return fmt.Errorf("invalid user profile")
	}
	_, err := h.Firestore.Collection("users").Doc(u.UID).Set(ctx, map[string]interface{}{
		"UID":           u.UID,
		"Username":      u.Username,
		"Email":         u.Email,
		"FirstName":     u.FirstName,
		"LastName":      u.LastName,
		"PhoneNumber":   u.PhoneNumber,
		"ProviderID":    u.ProviderID,
		"PhotoURL":      u.PhotoURL,
		"EmailVerified": u.EmailVerified,
		"CreatedAt":     u.CreatedAt,
		"UpdatedAt":     u.UpdatedAt,
	})
	return err
}

func (h *UserHandler) getProfile(ctx context.Context, uid string) (*models.User, error) {
	dsnap, err := h.Firestore.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m models.User
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}
