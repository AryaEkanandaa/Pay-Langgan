package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/models"
)

func Success(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessCreated(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessList(c echo.Context, message string, data interface{}, meta *models.Pagination) error {
	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Error(c echo.Context, status int, message string) error {
	return c.JSON(status, models.APIResponse{
		Success: false,
		Message: message,
	})
}

func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c echo.Context, message string) error {
	return Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c echo.Context, message string) error {
	return Error(c, http.StatusForbidden, message)
}

func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, message)
}

func Conflict(c echo.Context, message string) error {
	return Error(c, http.StatusConflict, message)
}

func InternalError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}
