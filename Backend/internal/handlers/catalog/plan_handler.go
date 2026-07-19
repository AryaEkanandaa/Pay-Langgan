package catalog

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services/catalog"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type PlanHandler struct {
	planService *catalog.PlanService
}

func NewPlanHandler(planService *catalog.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

func (h *PlanHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := utils.ParsePagination(c)

	plans, total, err := h.planService.GetAll(businessID, page, limit, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve plans")
	}

	return utils.SuccessList(c, "Plans retrieved successfully", plans, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *PlanHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid plan id")
	}

	plan, err := h.planService.GetByID(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "plan not found")
		}
		return utils.InternalError(c, "failed to retrieve plan")
	}

	return utils.Success(c, "Plan retrieved successfully", plan)
}

func (h *PlanHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.CreatePlanRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	plan, err := h.planService.Create(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "product not found")
		}
		return utils.InternalError(c, "failed to create plan")
	}

	return utils.SuccessCreated(c, "Plan created successfully", plan)
}

func (h *PlanHandler) Update(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid plan id")
	}

	var req models.UpdatePlanRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	plan, err := h.planService.Update(id, businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "plan not found")
		}
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		return utils.InternalError(c, "failed to update plan")
	}

	return utils.Success(c, "Plan updated successfully", plan)
}

func (h *PlanHandler) Delete(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid plan id")
	}

	err = h.planService.Delete(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "plan not found")
		}
		return utils.InternalError(c, "failed to delete plan")
	}

	return utils.Success(c, "Plan deleted successfully", nil)
}
