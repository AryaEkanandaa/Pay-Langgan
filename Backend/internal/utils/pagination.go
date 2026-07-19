package utils

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func ParsePagination(c echo.Context) (page, limit int, search string) {
	page = 1
	limit = 10

	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	search = c.QueryParam("search")
	return
}
