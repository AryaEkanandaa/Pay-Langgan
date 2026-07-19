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

type ProductHandler struct {
	productService *catalog.ProductService
}

func NewProductHandler(productService *catalog.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, search := utils.ParsePagination(c)

	products, total, err := h.productService.GetAll(businessID, page, limit, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve products")
	}

	return utils.SuccessList(c, "Products retrieved successfully", products, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *ProductHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid product id")
	}

	product, err := h.productService.GetByID(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "product not found")
		}
		return utils.InternalError(c, "failed to retrieve product")
	}

	return utils.Success(c, "Product retrieved successfully", product)
}

func (h *ProductHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	product, err := h.productService.Create(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, "name and service_id are required")
		}
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "service not found")
		}
		return utils.InternalError(c, "failed to create product")
	}

	return utils.SuccessCreated(c, "Product created successfully", product)
}

func (h *ProductHandler) Update(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid product id")
	}

	var req models.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	product, err := h.productService.Update(id, businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "product not found")
		}
		return utils.InternalError(c, "failed to update product")
	}

	return utils.Success(c, "Product updated successfully", product)
}

func (h *ProductHandler) Delete(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid product id")
	}

	err = h.productService.Delete(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "product not found")
		}
		return utils.InternalError(c, "failed to delete product")
	}

	return utils.Success(c, "Product deleted successfully", nil)
}
