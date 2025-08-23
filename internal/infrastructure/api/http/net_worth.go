package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/api/middleware"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type NetWorthHandler struct {
	service ports.NetWorthServicePort
	cfg     *config.Configuration
}

func NewNetWorthHandler(service ports.NetWorthServicePort, cfg *config.Configuration) *NetWorthHandler {
	if service == nil || cfg == nil {
		return nil
	}
	return &NetWorthHandler{service: service, cfg: cfg}
}

func (h *NetWorthHandler) GetNetWorth(c *gin.Context) {
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
		response.ErrorResponse(c, "unauthorized access to get net worth", nil, h.cfg.IsDevelopment())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	netWorth, err := h.service.GetNetWorth(c.Request.Context(), requestedUserID)
	if err != nil {
		response.ErrorResponse(c, "Failed to get net worth", err, h.cfg.IsDevelopment())
		return
	}
	response.SuccessResponseData(c, netWorth)
}
