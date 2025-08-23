package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type ExpenseHandler struct {
	expenseService ports.ExpenseServicePort
}

func NewExpenseHandler(expenseService ports.ExpenseServicePort) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) AddExpense(c *gin.Context) {
	userID := c.Param("id")
	if strings.TrimSpace(userID) == "" {
		response.ErrorResponse(c, "User ID is required", nil)
		return
	}

	var req dtos.AddExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "Invalid request body", err)
		return
	}

	input := ports.AddExpenseInput{
		UserID:              userID,
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
		response.ErrorResponse(c, "Failed to add expense", err)
		return
	}

	res := dtos.NewExpenseResponse(expense)
	response.SuccessWithStatusResponse(c, http.StatusCreated, "Expense added successfully", res)
}

func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	userID := c.Param("id")
	if strings.TrimSpace(userID) == "" {
		response.ErrorResponse(c, "User ID is required", nil)
		return
	}

	expenses, err := h.expenseService.ListExpenses(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResponse(c, "Failed to list expenses", err)
		return
	}

	res := dtos.NewListExpenseResponse(expenses)
	response.SuccessResponseData(c, res)
}

func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	userID := c.Param("id")
	expenseID := c.Param("expenseID")

	if strings.TrimSpace(userID) == "" {
		response.ErrorResponse(c, "User ID is required", nil)
		return
	}
	if strings.TrimSpace(expenseID) == "" {
		response.ErrorResponse(c, "Expense ID is required", nil)
		return
	}

	expense, err := h.expenseService.GetExpense(c.Request.Context(), userID, expenseID)
	if err != nil {
		response.ErrorResponse(c, "Failed to get expense", err)
		return
	}

	res := dtos.NewExpenseResponse(expense)
	response.SuccessResponseData(c, res)
}

func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	userID := c.Param("id")
	expenseID := c.Param("expenseID")

	if strings.TrimSpace(userID) == "" {
		response.ErrorResponse(c, "User ID is required", nil)
		return
	}
	if strings.TrimSpace(expenseID) == "" {
		response.ErrorResponse(c, "Expense ID is required", nil)
		return
	}

	var req dtos.AddExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "Invalid request body", err)
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

	expense, err := h.expenseService.UpdateExpense(c.Request.Context(), userID, expenseID, input)
	if err != nil {
		response.ErrorResponse(c, "Failed to update expense", err)
		return
	}

	res := dtos.NewExpenseResponse(expense)
	response.SuccessResponseData(c, res)
}

func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	userID := c.Param("id")
	expenseID := c.Param("expenseID")

	if strings.TrimSpace(userID) == "" {
		response.ErrorResponse(c, "User ID is required", nil)
		return
	}
	if strings.TrimSpace(expenseID) == "" {
		response.ErrorResponse(c, "Expense ID is required", nil)
		return
	}

	err := h.expenseService.DeleteExpense(c.Request.Context(), userID, expenseID)
	if err != nil {
		response.ErrorResponse(c, "Failed to delete expense", err)
		return
	}

	response.SuccessResponse(c, "Expense deleted successfully", gin.H{"user_id": userID, "income_id": expenseID})
}
