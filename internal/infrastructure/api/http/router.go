package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	middleware2 "github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
)

// NewRouter initializes the Gin engine, applies middleware, and registers all routes.
func NewRouter(healthHandler *HealthHandler, userService ports.UserServicePort, incomeService ports.IncomeServicePort, expenseService ports.ExpenseServicePort) *gin.Engine {
	router := gin.Default()

	router.Use(middleware2.CORS())
	router.Use(middleware2.RateLimit(60, time.Minute))

	registerHealthRoutes(router, healthHandler)

	// Register routes that depend on application services
	userHandler := NewUserHandler(userService)
	RegisterUserRoutes(router, userHandler)

	incomeHandler := NewIncomeHandler(incomeService)
	RegisterIncomeRoutes(router, incomeHandler)

	// Income Sources & processing
	incomeSourceHandler := NewIncomeSourceHandler(incomeService)
	RegisterIncomeSourceRoutes(router, incomeSourceHandler)

	// Expenses
	expenseHandler := NewExpenseHandler(expenseService)
	RegisterExpenseRoutes(router, expenseHandler)

	return router
}

func registerHealthRoutes(router *gin.Engine, healthHandler *HealthHandler) {
	router.GET("/health", healthHandler.HealthCheck)
}
