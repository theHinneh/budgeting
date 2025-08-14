package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/adapters/api/middleware"
	fbdb "github.com/theHinneh/budgeting/internal/adapters/db/firebase"
	"github.com/theHinneh/budgeting/internal/core/services"
)

// NewRouter initializes the Gin engine, applies middleware, and registers all routes.
func NewRouter(healthHandler *HealthHandler, fb *fbdb.Database) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(60, time.Minute))

	registerHealthRoutes(router, healthHandler)

	// Conditionally register user routes that depends on Firebase
	if fb != nil {
		svc := services.NewUserService(fb)
		userHandler := NewUserHandler(svc)
		RegisterUserRoutes(router, userHandler)
	}

	return router
}

func registerHealthRoutes(router *gin.Engine, healthHandler *HealthHandler) {
	router.GET("/health", healthHandler.HealthCheck)
}
