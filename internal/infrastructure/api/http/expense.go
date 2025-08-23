package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type ExpenseHandler struct {
	expenseService ports.ExpenseServicePort
	cfg            *config.Configuration
}

func NewExpenseHandler(expenseService ports.ExpenseServicePort, cfg *config.Configuration) *ExpenseHandler {
	if expenseService == nil || cfg == nil {
		return nil
	}
	return &ExpenseHandler{expenseService: expenseService, cfg: cfg}
}

func (h *ExpenseHandler) AddExpense(c *gin.Context) {
	requestedUserID := c.Param("id")
	if strings.TrimSpace(requestedUserID) == "" {
		response.ErrorResponse(c, "User ID is required", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to add expense", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req dtos.AddExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "Invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	input := ports.AddExpenseInput{
		UserID:              requestedUserID,
		Source:              req.ToDomain().Source,
		Amount:              req.ToDomain().Amount,
		Currency:            req.ToDomain().Currency,
		Notes:               req.ToDomain().Notes,
		IsRecurring:         req.ToDomain().IsRecurring,
		RecurrenceFrequency: req.ToDomain().RecurrenceFrequency,
		NextOccurrenceDate:  &req.ToDomain().NextOccurrenceDate,
	}

	expense, err := h.expenseService.AddExpense(c.Request.Context(), input)
	if err != nil {
		response.ErrorResponse(c, "Failed to add expense", err, h.cfg.IsDevelopment())
		return
	}

	res := dtos.NewExpenseResponse(expense)
	response.SuccessWithStatusResponse(c, http.StatusCreated, "Expense added successfully", res)
}

func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	requestedUserID := c.Param("id")
	if strings.TrimSpace(requestedUserID) == "" {
		response.ErrorResponse(c, "User ID is required", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to list expenses", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	expenses, err := h.expenseService.ListExpenses(c.Request.Context(), requestedUserID)
	if err != nil {
		response.ErrorResponse(c, "Failed to list expenses", err, h.cfg.IsDevelopment())
		return
	}

	res := dtos.NewListExpenseResponse(expenses)
	response.SuccessResponseData(c, res)
}

func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	requestedUserID := c.Param("id")
	expenseID := c.Param("expenseID")

	if strings.TrimSpace(requestedUserID) == "" {
		response.ErrorResponse(c, "User ID is required", nil, h.cfg.IsDevelopment())
		return
	}
	if strings.TrimSpace(expenseID) == "" {
		response.ErrorResponse(c, "Expense ID is required", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to get expense", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	expense, err := h.expenseService.GetExpense(c.Request.Context(), requestedUserID, expenseID)
	if err != nil {
		response.ErrorResponse(c, "Failed to get expense", err, h.cfg.IsDevelopment())
		return
	}

	res := dtos.NewExpenseResponse(expense)
	response.SuccessResponseData(c, res)
}

func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	requestedUserID := c.Param("id")
	expenseID := c.Param("expenseID")

	if strings.TrimSpace(requestedUserID) == "" {
		response.ErrorResponse(c, "User ID is required", nil, h.cfg.IsDevelopment())
		return
	}
	if strings.TrimSpace(expenseID) == "" {
		response.ErrorResponse(c, "Expense ID is required", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to update expense", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req dtos.AddExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "Invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	input := ports.AddExpenseInput{
		Source:              req.ToDomain().Source,
		Amount:              req.ToDomain().Amount,
		Currency:            req.ToDomain().Currency,
		Notes:               req.ToDomain().Notes,
		IsRecurring:         req.ToDomain().IsRecurring,
		RecurrenceFrequency: req.ToDomain().RecurrenceFrequency,
		NextOccurrenceDate:  &req.ToDomain().NextOccurrenceDate,
	}

	expense, err := h.expenseService.UpdateExpense(c.Request.Context(), requestedUserID, expenseID, input)
	if err != nil {
		response.ErrorResponse(c, "Failed to update expense", err, h.cfg.IsDevelopment())
		return
	}

	res := dtos.NewExpenseResponse(expense)
	response.SuccessResponseData(c, res)
}

func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	requestedUserID := c.Param("id")
	expenseID := c.Param("expenseID")

	if strings.TrimSpace(requestedUserID) == "" {
		response.ErrorResponse(c, "User ID is required", nil, h.cfg.IsDevelopment())
		return
	}
	if strings.TrimSpace(expenseID) == "" {
		response.ErrorResponse(c, "Expense ID is required", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to delete expense", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err := h.expenseService.DeleteExpense(c.Request.Context(), requestedUserID, expenseID)
	if err != nil {
		response.ErrorResponse(c, "Failed to delete expense", err, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponse(c, "Expense deleted successfully", gin.H{"user_id": requestedUserID, "income_id": expenseID})
}
