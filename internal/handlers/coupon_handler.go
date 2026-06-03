package handlers

import (
	"errors"
	"strconv"

	"pay-langgan/internal/models"
	"pay-langgan/internal/services"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type CouponHandler struct {
	couponService *services.CouponService
}

func NewCouponHandler(couponService *services.CouponService) *CouponHandler {
	return &CouponHandler{couponService: couponService}
}

func (h *CouponHandler) List(c echo.Context) error {
	page, limit, search := parsePagination(c)

	coupons, total, err := h.couponService.GetAll(page, limit, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve coupons")
	}

	return utils.SuccessList(c, "Coupons retrieved successfully", coupons, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *CouponHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid coupon id")
	}

	coupon, err := h.couponService.GetByID(id)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "coupon not found")
		}
		return utils.InternalError(c, "failed to retrieve coupon")
	}

	return utils.Success(c, "Coupon retrieved successfully", coupon)
}

func (h *CouponHandler) Create(c echo.Context) error {
	var req models.CreateCouponRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	coupon, err := h.couponService.Create(req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		if errors.Is(err, utils.ErrConflict) {
			return utils.Conflict(c, "coupon code already exists")
		}
		return utils.InternalError(c, "failed to create coupon")
	}

	return utils.SuccessCreated(c, "Coupon created successfully", coupon)
}

func (h *CouponHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid coupon id")
	}

	var req models.UpdateCouponRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	coupon, err := h.couponService.Update(id, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "coupon not found")
		}
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, err.Error())
		}
		if errors.Is(err, utils.ErrConflict) {
			return utils.Conflict(c, "coupon code already exists")
		}
		return utils.InternalError(c, "failed to update coupon")
	}

	return utils.Success(c, "Coupon updated successfully", coupon)
}

func (h *CouponHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid coupon id")
	}

	err = h.couponService.Delete(id)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "coupon not found")
		}
		return utils.InternalError(c, "failed to delete coupon")
	}

	return utils.Success(c, "Coupon deleted successfully", nil)
}
