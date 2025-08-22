package http

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type HealthHandler struct {
	config          *config.Configuration
	firestoreClient *firestore.Client
}

func NewHealthHandler(cfg *config.Configuration, firestoreClient *firestore.Client) *HealthHandler {
	return &HealthHandler{
		config:          cfg,
		firestoreClient: firestoreClient,
	}
}

func (repo *HealthHandler) HealthCheck(ctx *gin.Context) {
	healthData := gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	}

	env := "unknown"
	if repo.config != nil && repo.config.V != nil {
		env = repo.config.V.GetString("APP_ENV")
		if env == "" {
			env = repo.config.V.GetString("app_env")
		}
	}
	healthData["environment"] = env

	if repo.firestoreClient != nil {
		// Perform a simple Firestore operation to check connectivity
		_, err := repo.firestoreClient.Collection("health_check").Doc("status").Get(ctx.Request.Context())
		if err != nil && err.Error() != "rpc error: code = NotFound desc = document not found: projects/budgeting-bb0f5/databases/(default)/documents/health_check/status" {
			healthData["database"] = gin.H{
				"status":  "unhealthy",
				"details": err.Error(),
			}
		} else {
			healthData["database"] = gin.H{"status": "healthy"}
		}
	}

	response.SuccessResponse(ctx, "Health check successful", healthData)
}
