package payment

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	service "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
)

func (h *CheckoutHandler) History(c echo.Context) error {
	page, limit, _ := utils.ParsePagination(c)
	ids := []int{0, 0}
	for i, key := range []string{"invoice_id", "subscription_id"} {
		if value := c.QueryParam(key); value != "" {
			id, err := strconv.Atoi(value)
			if err != nil || id < 1 {
				return utils.BadRequest(c, "invalid "+key)
			}
			ids[i] = id
		}
	}
	items, total, err := h.service.PaymentHistory(c.Request().Context(), middlewares.GetBusinessID(c), ids[0], ids[1], page, limit, c.QueryParam("status"))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.SuccessList(c, "Payment history retrieved", items, &models.Pagination{Page: page, Limit: limit, Total: total})
}

func (h *CheckoutHandler) Methods(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid customer id")
	}
	items, err := h.service.PaymentMethods(c.Request().Context(), middlewares.GetBusinessID(c), id)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Payment methods retrieved", items)
}

func (h *CheckoutHandler) GetAutopay(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid subscription id")
	}
	settings, err := h.service.AutopaySettings(c.Request().Context(), id, middlewares.GetBusinessID(c))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Autopay settings retrieved", settings)
}

func (h *CheckoutHandler) SetAutopay(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid subscription id")
	}
	var request service.AutopayRequest
	if err := c.Bind(&request); err != nil {
		return utils.BadRequest(c, "invalid autopay request")
	}
	settings, err := h.service.SetAutopay(c.Request().Context(), id, middlewares.GetBusinessID(c), middlewares.GetUserID(c), request)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Autopay settings updated", settings)
}
