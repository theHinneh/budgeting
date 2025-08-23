package http

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/response"
)

type NetWorthHandler struct {
	service ports.NetWorthServicePort
}

func NewNetWorthHandler(service ports.NetWorthServicePort) *NetWorthHandler {
	return &NetWorthHandler{service: service}
}

func (h *NetWorthHandler) GetNetWorth(c *gin.Context) {
	userID := c.Param("id")
	if strings.TrimSpace(userID) == "" {
		response.ErrorResponse(c, "User ID is required", nil)
		return
	}

	netWorth, err := h.service.GetNetWorth(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResponse(c, "Failed to get net worth", err)
		return
	}
	response.SuccessResponseData(c, netWorth)
}
