package http

import (
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/core/ports"
	"github.com/theHinneh/budgeting/pkg/config"
	"github.com/theHinneh/budgeting/pkg/logger"
	"github.com/theHinneh/budgeting/pkg/response"
	"go.uber.org/zap"
)

type HealthHandler struct {
	Cfg *config.Configuration
	DB  ports.DatabasePort
}

func NewHealthHandler(cfg *config.Configuration, db ports.DatabasePort) *HealthHandler {
	return &HealthHandler{Cfg: cfg, DB: db}
}

func (repo *HealthHandler) HealthCheck(ctx *gin.Context) {
	logger.Info("Health check request received")

	// Safely read environment to avoid nil pointer if Cfg is not set
	env := "unknown"
	if repo.Cfg != nil && repo.Cfg.V != nil {
		env = repo.Cfg.V.GetString("APP_ENV")
		if env == "" {
			env = repo.Cfg.V.GetString("app_env")
		}
	}

	healthData := gin.H{
		"status":      "healthy",
		"environment": env,
	}

	if repo.DB != nil {
		if err := repo.DB.Ping(ctx.Request.Context()); err != nil {
			healthData["database"] = gin.H{
				"status":  "unhealthy",
				"details": err.Error(),
			}
			healthData["status"] = "unhealthy"
			logger.Error("Database health check failed", zap.Error(err))
		} else {
			healthData["database"] = gin.H{
				"status": "healthy",
			}
		}
	} else {
		healthData["database"] = gin.H{
			"status":  "not_initialized",
			"details": "Database connection not initialized",
		}
		logger.Error("Database not initialized during health check")
	}

	response.SuccessResponseData(ctx, healthData)
}
