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

type IncomeHandler struct {
	Service ports.IncomeServicePort
	cfg     *config.Configuration
}

func NewIncomeHandler(svc ports.IncomeServicePort, cfg *config.Configuration) *IncomeHandler {
	if svc == nil || cfg == nil {
		return nil
	}
	return &IncomeHandler{Service: svc, cfg: cfg}
}

func (h *IncomeHandler) AddIncome(c *gin.Context) {
	requestedUserID := strings.TrimSpace(c.Param("id"))
	if requestedUserID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to add income", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req dtos.AddIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}
	income, err := h.Service.AddIncome(c.Request.Context(), ports.AddIncomeInput{
		UserID:   requestedUserID,
		Source:   req.ToDomain().Source,
		Amount:   req.ToDomain().Amount,
		Currency: req.ToDomain().Currency,
		Notes:    req.ToDomain().Notes,
	})
	if err != nil {
		response.ErrorResponse(c, "failed to add income", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessWithStatusResponse(c, http.StatusCreated, "income added", income)
}

func (h *IncomeHandler) AddIncomeSource(c *gin.Context) {
	requestedUserID := strings.TrimSpace(c.Param("id"))
	//layout := "2006-12-31"

	if requestedUserID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to add income source", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req dtos.AddIncomeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err, h.cfg.IsDevelopment())
		return
	}

	//parsedTime, err := time.Parse(layout, req.NextPayAt)

	src, err := h.Service.AddIncomeSource(c.Request.Context(), ports.AddIncomeSourceInput{
		UserID:    requestedUserID,
		Source:    req.Source,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Frequency: ports.PayFrequency(req.Frequency),
		NextPayAt: req.NextPayAt,
		Notes:     req.Notes,
	})
	if err != nil {
		response.ErrorResponse(c, "failed to add income source", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessWithStatusResponse(c, http.StatusCreated, "income source created", src)
}

func (h *IncomeHandler) ListIncomes(c *gin.Context) {
	requestedUserID := strings.TrimSpace(c.Param("id"))
	if requestedUserID == "" {
		response.ErrorResponse(c, "missing user id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to list incomes", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	incomes, err := h.Service.ListIncomes(c.Request.Context(), requestedUserID)
	if err != nil {
		response.ErrorResponse(c, "failed to list incomes", err, h.cfg.IsDevelopment())
		return
	}

	if len(incomes) == 0 {
		response.ErrorResponse(c, "no incomes found", nil, h.cfg.IsDevelopment())
		return
	}

	response.SuccessResponseData(c, incomes)
}

func (h *IncomeHandler) DeleteIncome(c *gin.Context) {
	requestedUserID := strings.TrimSpace(c.Param("id"))
	incomeID := strings.TrimSpace(c.Param("incomeId"))
	if requestedUserID == "" || incomeID == "" {
		response.ErrorResponse(c, "missing user id or income id", nil, h.cfg.IsDevelopment())
		return
	}

	authUID, exists := c.Get(middleware.FirebaseUIDKey)
	if !exists {
		response.ErrorResponse(c, "authenticated user ID not found in context", nil, h.cfg.IsDevelopment())
		return
	}

	if requestedUserID != authUID.(string) {
		response.ErrorResponse(c, "unauthorized access to delete income", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := h.Service.DeleteIncome(c.Request.Context(), requestedUserID, incomeID); err != nil {
		response.ErrorResponse(c, "failed to delete income", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponse(c, "income deleted", gin.H{"user_id": requestedUserID, "income_id": incomeID})
}
