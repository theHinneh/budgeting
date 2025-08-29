package http

import (
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	middleware2 "github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
)

func NewRouter(
	healthHandler *HealthHandler, userService ports.UserServicePort, incomeService ports.IncomeServicePort,
	expenseService ports.ExpenseServicePort, netWorthService ports.NetWorthServicePort, firebaseApp *firebase.App,
	authService ports.AuthServicePort, cfg *config.Configuration,
) *gin.Engine {
	router := gin.Default()

	serverConfig := cfg.GetServerConfig()
	router.Use(middleware2.CORS(serverConfig.CORSOrigin))
	router.Use(middleware2.RateLimit(60, time.Minute))

	registerHealthRoutes(router, healthHandler)

	userHandler := NewUserHandler(userService, firebaseApp, cfg)
	authHandler := NewAuthHandler(firebaseApp, authService, cfg)

	publicV1 := router.Group("/v1")
	{
		publicV1.POST("/users", userHandler.CreateUser)
		publicV1.POST("/auth/login", authHandler.Login)
		publicV1.POST("/auth/refresh", authHandler.RefreshToken)
		publicV1.POST("/auth/forgot-password", userHandler.ForgotPassword)
	}

	v1 := router.Group("/v1")
	v1.Use(middleware2.FirebaseAuthentication(firebaseApp, cfg))
	{

		userRoutes := v1.Group("/users")
		{
			userRoutes.GET("/:id", userHandler.GetUser)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.DELETE("/:id", userHandler.DeleteUser)
			userRoutes.POST("/:id/password", userHandler.ChangePassword)
		}

		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/logout", authHandler.Logout)
			authRoutes.GET("/me", authHandler.GetCurrentUser)
			authRoutes.GET("/sessions", authHandler.GetUserSessions)
			authRoutes.POST("/sessions/revoke", authHandler.RevokeSession)
			authRoutes.POST("/sessions/revoke-all", authHandler.RevokeAllSessions)
		}

		incomeHandler := NewIncomeHandler(incomeService, cfg)
		incomeRoutes := v1.Group("/users/:id/incomes")
		{
			incomeRoutes.POST("", incomeHandler.AddIncome)
			incomeRoutes.GET("", incomeHandler.ListIncomes)
			incomeRoutes.DELETE(":incomeId", incomeHandler.DeleteIncome)
		}

		incomeSourceHandler := NewIncomeSourceHandler(incomeService, cfg)
		incomeSourceRoutes := v1.Group("/users/:id")
		{
			incomeSourceRoutes.POST("/income-sources", incomeSourceHandler.AddIncomeSource)
			incomeSourceRoutes.GET("/income-sources", incomeSourceHandler.ListIncomeSources)
			incomeSourceRoutes.POST("/incomes/process-due", incomeSourceHandler.ProcessDueIncomes)
		}

		expenseHandler := NewExpenseHandler(expenseService, cfg)
		expenseRoutes := v1.Group("/users/:id/expenses")
		{
			expenseRoutes.POST("", expenseHandler.AddExpense)
			expenseRoutes.GET("", expenseHandler.ListExpenses)
			expenseRoutes.GET("/:expenseID", expenseHandler.GetExpense)
			expenseRoutes.PUT("/:expenseID", expenseHandler.UpdateExpense)
			expenseRoutes.DELETE("/:expenseID", expenseHandler.DeleteExpense)
		}

		netWorthHandler := NewNetWorthHandler(netWorthService, cfg)
		netWorthRoutes := v1.Group("/users/:id/net-worth")
		{
			netWorthRoutes.GET("", netWorthHandler.GetNetWorth)
		}
	}

	return router
}

func registerHealthRoutes(router *gin.Engine, healthHandler *HealthHandler) {
	router.GET("/health", healthHandler.HealthCheck)
}
