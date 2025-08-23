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

	v1 := router.Group("/v1")
	{
		userHandler := NewUserHandler(userService)
		userRoutes := v1.Group("/users")
		{
			userRoutes.POST("", userHandler.CreateUser)
			userRoutes.GET("/:id", userHandler.GetUser)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.DELETE("/:id", userHandler.DeleteUser)
			userRoutes.POST("/:id/password", userHandler.ChangePassword)
		}
		v1.POST("/auth/forgot-password", userHandler.ForgotPassword)

		incomeHandler := NewIncomeHandler(incomeService)
		incomeRoutes := v1.Group("/users/:id/incomes")
		{
			incomeRoutes.POST("", incomeHandler.AddIncome)
			incomeRoutes.GET("", incomeHandler.ListIncomes)
			incomeRoutes.DELETE(":incomeId", incomeHandler.DeleteIncome)
		}

		incomeSourceHandler := NewIncomeSourceHandler(incomeService)
		incomeSourceRoutes := v1.Group("/users/:id")
		{
			incomeSourceRoutes.POST("/income-sources", incomeSourceHandler.AddIncomeSource)
			incomeSourceRoutes.GET("/income-sources", incomeSourceHandler.ListIncomeSources)
			incomeSourceRoutes.POST("/incomes/process-due", incomeSourceHandler.ProcessDueIncomes)
		}

		expenseHandler := NewExpenseHandler(expenseService)
		expenseRoutes := v1.Group("/users/:id/expenses")
		{
			expenseRoutes.POST("", expenseHandler.AddExpense)
			expenseRoutes.GET("", expenseHandler.ListExpenses)
			expenseRoutes.GET("/:expenseID", expenseHandler.GetExpense)
			expenseRoutes.PUT("/:expenseID", expenseHandler.UpdateExpense)
			expenseRoutes.DELETE("/:expenseID", expenseHandler.DeleteExpense)
		}
	}

	return router
}

func registerHealthRoutes(router *gin.Engine, healthHandler *HealthHandler) {
	router.GET("/health", healthHandler.HealthCheck)
}
