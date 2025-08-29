package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/dto"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type IncomeSourceHandler struct {
	Service ports.IncomeServicePort
	cfg     *config.Configuration
}

func NewIncomeSourceHandler(svc ports.IncomeServicePort, cfg *config.Configuration) *IncomeSourceHandler {
	if svc == nil || cfg == nil {
		return nil
	}
	return &IncomeSourceHandler{Service: svc, cfg: cfg}
}

func (h *IncomeSourceHandler) AddIncomeSource(c *gin.Context) {
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

	src, err := h.Service.AddIncomeSource(c.Request.Context(), dto.AddIncomeSourceInput{
		UserID:    requestedUserID,
		Source:    req.Source,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Frequency: dto.PayFrequency(req.Frequency),
		NextPayAt: req.NextPayAt,
		Notes:     req.Notes,
	})
	if err != nil {
		response.ErrorResponse(c, "failed to add income source", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessWithStatusResponse(c, http.StatusCreated, "income source created", src)
}

func (h *IncomeSourceHandler) ListIncomeSources(c *gin.Context) {
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
		response.ErrorResponse(c, "unauthorized access to list income sources", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	sources, err := h.Service.ListIncomeSources(c.Request.Context(), requestedUserID)
	if err != nil {
		response.ErrorResponse(c, "failed to list income sources", err, h.cfg.IsDevelopment())
		return
	}

	if len(sources) == 0 {
		response.ErrorResponse(c, "no income sources found", nil, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponseData(c, sources)
}

func (h *IncomeSourceHandler) ProcessDueIncomes(c *gin.Context) {
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
		response.ErrorResponse(c, "unauthorized access to process due incomes", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	count, err := h.Service.ProcessDueIncomes(c.Request.Context(), requestedUserID, time.Now())
	if err != nil {
		response.ErrorResponse(c, "failed to process due incomes", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponse(c, "processed due incomes", gin.H{"created": count})
}
