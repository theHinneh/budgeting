package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/adapters/api/middleware"
)

func NewRouter(healthHandler *HealthHandler) *gin.Engine {
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS())

	// Apply rate limiting to all routes - 60 requests per minute
	router.Use(middleware.RateLimit(60, time.Minute))

	// Register routes from modules
	registerHealthRoutes(router, healthHandler)
	return router
}

func registerHealthRoutes(router *gin.Engine, healthHandler *HealthHandler) {
	router.GET("health", healthHandler.HealthCheck)
}
