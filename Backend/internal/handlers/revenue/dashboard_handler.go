package revenue

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	revenuesvc "pay-langgan/internal/services/revenue"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	dashboardService *revenuesvc.DashboardService
}

func NewDashboardHandler(dashboardService *revenuesvc.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) Summary(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	summary, err := h.dashboardService.GetSummary(businessID, models.Role(middlewares.GetRole(c)))
	if err != nil {
		return utils.InternalError(c, "failed to retrieve dashboard summary")
	}

	return utils.Success(c, "Dashboard summary retrieved successfully", summary)
}
