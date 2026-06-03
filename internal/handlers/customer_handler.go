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

type CustomerHandler struct {
	customerService *services.CustomerService
}

func NewCustomerHandler(customerService *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

func (h *CustomerHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := parsePagination(c)

	customers, total, err := h.customerService.GetAll(businessID, page, limit, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve customers")
	}

	return utils.SuccessList(c, "Customers retrieved successfully", customers, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *CustomerHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid customer id")
	}

	customer, err := h.customerService.GetByID(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "customer not found")
		}
		return utils.InternalError(c, "failed to retrieve customer")
	}

	return utils.Success(c, "Customer retrieved successfully", customer)
}

func (h *CustomerHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.CreateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	customer, err := h.customerService.Create(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, "name is required")
		}
		return utils.InternalError(c, "failed to create customer")
	}

	return utils.SuccessCreated(c, "Customer created successfully", customer)
}

func (h *CustomerHandler) Update(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid customer id")
	}

	var req models.UpdateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	customer, err := h.customerService.Update(id, businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "customer not found")
		}
		return utils.InternalError(c, "failed to update customer")
	}

	return utils.Success(c, "Customer updated successfully", customer)
}

func (h *CustomerHandler) Delete(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid customer id")
	}

	err = h.customerService.Delete(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "customer not found")
		}
		return utils.InternalError(c, "failed to delete customer")
	}

	return utils.Success(c, "Customer deleted successfully", nil)
}
