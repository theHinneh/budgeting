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

	// Conditionally register routes that depend on Firebase
	if fb != nil {
		// Users
		userSvc := services.NewUserService(fb)
		userHandler := NewUserHandler(userSvc)
		RegisterUserRoutes(router, userHandler)

		// Incomes
		incomeRepo := fb
		incomeSvc := services.NewIncomeService(incomeRepo)
		incomeHandler := NewIncomeHandler(incomeSvc)
		RegisterIncomeRoutes(router, incomeHandler)

		// Income Sources & processing
		incomeSourceHandler := NewIncomeSourceHandler(incomeSvc)
		RegisterIncomeSourceRoutes(router, incomeSourceHandler)
	}

	return router
}

func registerHealthRoutes(router *gin.Engine, healthHandler *HealthHandler) {
	router.GET("/health", healthHandler.HealthCheck)
}
