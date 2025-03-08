package pagination

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// PaginationParams параметры пагинации
type PaginationParams struct {
	Page  int
	Limit int
}

// ExtractPaginationParams извлекает параметры пагинации из запроса
func ExtractPaginationParams(c echo.Context) PaginationParams {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}
