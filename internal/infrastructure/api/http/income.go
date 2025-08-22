package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type IncomeHandler struct {
	Service ports.IncomeServicePort
}

func NewIncomeHandler(svc ports.IncomeServicePort) *IncomeHandler {
	if svc == nil {
		return nil
	}
	return &IncomeHandler{Service: svc}
}

// RegisterIncomeRoutes attaches income endpoints under /users/:id/incomes
func RegisterIncomeRoutes(router *gin.Engine, ih *IncomeHandler) {
	if router == nil || ih == nil {
		return
	}
	g := router.Group("/users/:id/incomes")
	{
		g.POST("", ih.AddIncome)
		g.GET("", ih.ListIncomes)
		g.DELETE(":incomeId", ih.DeleteIncome)
	}
}

func (h *IncomeHandler) AddIncome(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}
	var req dtos.AddIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err)
		return
	}
	income, err := h.Service.AddIncome(c.Request.Context(), ports.AddIncomeInput{
		UserID:   userID,
		Source:   req.Source,
		Amount:   req.Amount,
		Currency: req.Currency,
		Notes:    req.Notes,
	})
	if err != nil {
		response.ErrorResponse(c, "failed to add income", err)
		return
	}
	response.SuccessWithStatusResponse(c, http.StatusCreated, "income added", income)
}

func (h *IncomeHandler) ListIncomes(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}
	incomes, err := h.Service.ListIncomes(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResponse(c, "failed to list incomes", err)
		return
	}

	if len(incomes) == 0 {
		response.ErrorResponse(c, "no incomes found", nil)
		return
	}

	response.SuccessResponseData(c, incomes)
}

func (h *IncomeHandler) DeleteIncome(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	incomeID := strings.TrimSpace(c.Param("incomeId"))
	if userID == "" || incomeID == "" {
		response.ErrorResponse(c, "missing user id or income id", nil)
		return
	}
	if err := h.Service.DeleteIncome(c.Request.Context(), userID, incomeID); err != nil {
		response.ErrorResponse(c, "failed to delete income", err)
		return
	}
	response.SuccessResponse(c, "income deleted", gin.H{"user_id": userID, "income_id": incomeID})
}
