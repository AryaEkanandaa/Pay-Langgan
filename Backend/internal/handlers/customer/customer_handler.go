package customer

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services/customer"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type CustomerHandler struct {
	customerService *customer.CustomerService
}

func NewCustomerHandler(customerService *customer.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

func (h *CustomerHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := utils.ParsePagination(c)

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

	cust, err := h.customerService.GetByID(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "customer not found")
		}
		return utils.InternalError(c, "failed to retrieve customer")
	}

	return utils.Success(c, "Customer retrieved successfully", cust)
}

func (h *CustomerHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.CreateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	cust, err := h.customerService.Create(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		return utils.InternalError(c, "failed to create customer")
	}

	return utils.SuccessCreated(c, "Customer created successfully", cust)
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

	cust, err := h.customerService.Update(id, businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "customer not found")
		}
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		return utils.InternalError(c, "failed to update customer")
	}

	return utils.Success(c, "Customer updated successfully", cust)
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
