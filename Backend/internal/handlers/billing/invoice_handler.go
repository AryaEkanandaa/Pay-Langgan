package billing

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	billingservice "pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type InvoiceHandler struct {
	service *billingservice.InvoiceService
}

func NewInvoiceHandler(service *billingservice.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: service}
}

func (h *InvoiceHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := utils.ParsePagination(c)
	status := c.QueryParam("status")
	if status != "" && status != "pending" && status != "paid" && status != "failed" && status != "cancelled" {
		return utils.BadRequest(c, "invalid invoice status filter")
	}

	invoices, total, err := h.service.List(businessID, page, limit, status, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve invoices")
	}
	return utils.SuccessList(c, "Invoices retrieved successfully", invoices, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *InvoiceHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid invoice id")
	}

	invoice, err := h.service.GetDetail(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "invoice not found")
		}
		return utils.InternalError(c, "failed to retrieve invoice")
	}
	return utils.Success(c, "Invoice retrieved successfully", invoice)
}

func (h *InvoiceHandler) MarkPaid(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid invoice id")
	}

	invoice, err := h.service.MarkPaid(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "invoice not found")
		}
		return utils.BadRequest(c, err.Error())
	}
	return utils.Success(c, "Invoice marked as paid", invoice)
}
