package revenue

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	revenuesvc "pay-langgan/internal/services/revenue"
	"pay-langgan/internal/utils"
	"time"

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
	if c.QueryParam("from") != "" || c.QueryParam("to") != "" {
		location, err := h.dashboardService.RangeLocation(businessID)
		if err != nil {
			return utils.InternalError(c, "failed to retrieve business timezone")
		}
		granularity := c.QueryParam("granularity")
		if granularity == "" {
			granularity = "day"
		}
		if _, _, err = revenuesvc.ParseDashboardRange(c.QueryParam("from"), c.QueryParam("to"), granularity, location, time.Now()); err != nil {
			return utils.BadRequest(c, err.Error())
		}
		summary, err := h.dashboardService.GetRangeSummary(businessID, c.QueryParam("from"), c.QueryParam("to"), granularity, location, models.Role(middlewares.GetRole(c)))
		if err != nil {
			return utils.InternalError(c, "failed to retrieve dashboard summary")
		}
		return utils.Success(c, "Dashboard summary retrieved successfully", summary)
	}
	summary, err := h.dashboardService.GetSummary(businessID, models.Role(middlewares.GetRole(c)))
	if err != nil {
		return utils.InternalError(c, "failed to retrieve dashboard summary")
	}

	return utils.Success(c, "Dashboard summary retrieved successfully", summary)
}

func (h *DashboardHandler) PlatformSummary(c echo.Context) error {
	summary, err := h.dashboardService.GetPlatformSummary()
	if err != nil {
		return utils.InternalError(c, "failed to retrieve platform dashboard")
	}
	return utils.Success(c, "Platform dashboard retrieved successfully", summary)
}
