package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/dtos"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type IncomeSourceHandler struct {
	Service ports.IncomeServicePort
}

func NewIncomeSourceHandler(svc ports.IncomeServicePort) *IncomeSourceHandler {
	if svc == nil {
		return nil
	}
	return &IncomeSourceHandler{Service: svc}
}

func (h *IncomeSourceHandler) AddIncomeSource(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	//layout := "2006-12-31"

	if userID == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}
	var req dtos.AddIncomeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, "invalid request body", err)
		return
	}

	//parsedTime, err := time.Parse(layout, req.NextPayAt)

	src, err := h.Service.AddIncomeSource(c.Request.Context(), ports.AddIncomeSourceInput{
		UserID:    userID,
		Source:    req.Source,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Frequency: ports.PayFrequency(req.Frequency),
		NextPayAt: req.NextPayAt,
		Notes:     req.Notes,
	})
	if err != nil {
		response.ErrorResponse(c, "failed to add income source", err)
		return
	}
	response.SuccessWithStatusResponse(c, http.StatusCreated, "income source created", src)
}

func (h *IncomeSourceHandler) ListIncomeSources(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}
	sources, err := h.Service.ListIncomeSources(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResponse(c, "failed to list income sources", err)
		return
	}

	if len(sources) == 0 {
		response.ErrorResponse(c, "no income sources found", nil)
		return
	}
	response.SuccessResponseData(c, sources)
}

func (h *IncomeSourceHandler) ProcessDueIncomes(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		response.ErrorResponse(c, "missing user id", nil)
		return
	}
	// optional override time? For now, use server time
	count, err := h.Service.ProcessDueIncomes(c.Request.Context(), userID, time.Now())
	if err != nil {
		response.ErrorResponse(c, "failed to process due incomes", err)
		return
	}
	response.SuccessResponse(c, "processed due incomes", gin.H{"created": count})
}
