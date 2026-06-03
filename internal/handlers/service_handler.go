package handlers

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type ServiceHandler struct {
	serviceService *services.ServiceService
}

func NewServiceHandler(serviceService *services.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

func (h *ServiceHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := parsePagination(c)

	services, total, err := h.serviceService.GetAll(businessID, page, limit, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve services")
	}

	return utils.SuccessList(c, "Services retrieved successfully", services, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *ServiceHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid service id")
	}

	service, err := h.serviceService.GetByID(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "service not found")
		}
		return utils.InternalError(c, "failed to retrieve service")
	}

	return utils.Success(c, "Service retrieved successfully", service)
}

func (h *ServiceHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.CreateServiceRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	service, err := h.serviceService.Create(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, "name is required")
		}
		return utils.InternalError(c, "failed to create service")
	}

	return utils.SuccessCreated(c, "Service created successfully", service)
}

func (h *ServiceHandler) Update(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid service id")
	}

	var req models.UpdateServiceRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	service, err := h.serviceService.Update(id, businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "service not found")
		}
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		return utils.InternalError(c, "failed to update service")
	}

	return utils.Success(c, "Service updated successfully", service)
}

func (h *ServiceHandler) Delete(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid service id")
	}

	err = h.serviceService.Delete(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "service not found")
		}
		return utils.InternalError(c, "failed to delete service")
	}

	return utils.Success(c, "Service deleted successfully", nil)
}
