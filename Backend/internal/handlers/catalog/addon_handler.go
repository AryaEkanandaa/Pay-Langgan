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

type AddOnHandler struct {
	addOnService *catalog.AddOnService
}

func NewAddOnHandler(addOnService *catalog.AddOnService) *AddOnHandler {
	return &AddOnHandler{addOnService: addOnService}
}

func (h *AddOnHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := utils.ParsePagination(c)

	addOns, total, err := h.addOnService.GetAll(businessID, page, limit, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve add-ons")
	}

	return utils.SuccessList(c, "Add-ons retrieved successfully", addOns, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *AddOnHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid add-on id")
	}

	addOn, err := h.addOnService.GetByID(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "add-on not found")
		}
		return utils.InternalError(c, "failed to retrieve add-on")
	}

	return utils.Success(c, "Add-on retrieved successfully", addOn)
}

func (h *AddOnHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.CreateAddOnRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	addOn, err := h.addOnService.Create(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "product not found")
		}
		return utils.InternalError(c, "failed to create add-on")
	}

	return utils.SuccessCreated(c, "Add-on created successfully", addOn)
}

func (h *AddOnHandler) Update(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid add-on id")
	}

	var req models.UpdateAddOnRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	addOn, err := h.addOnService.Update(id, businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "add-on not found")
		}
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		return utils.InternalError(c, "failed to update add-on")
	}

	return utils.Success(c, "Add-on updated successfully", addOn)
}

func (h *AddOnHandler) Delete(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid add-on id")
	}

	err = h.addOnService.Delete(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "add-on not found")
		}
		return utils.InternalError(c, "failed to delete add-on")
	}

	return utils.Success(c, "Add-on deleted successfully", nil)
}
