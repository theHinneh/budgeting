package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application"
	middleware2 "github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
	fbdb "github.com/theHinneh/budgeting/internal/infrastructure/db/firebase"
)

// NewRouter initializes the Gin engine, applies middleware, and registers all routes.
func NewRouter(healthHandler *HealthHandler, fb *fbdb.Database) *gin.Engine {
	router := gin.Default()

	router.Use(middleware2.CORS())
	router.Use(middleware2.RateLimit(60, time.Minute))

	registerHealthRoutes(router, healthHandler)

	// Conditionally register routes that depend on Firebase
	if fb != nil {
		// Users
		userSvc := application.NewUserService(fb)
		userHandler := NewUserHandler(userSvc)
		RegisterUserRoutes(router, userHandler)

		// Incomes
		incomeRepo := fb
		incomeSvc := application.NewIncomeService(incomeRepo)
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
