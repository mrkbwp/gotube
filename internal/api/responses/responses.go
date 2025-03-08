// internal/api/responses/responses.go
package responses

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func Error(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

func JSON(c echo.Context, code int, data interface{}) error {
	return c.JSON(code, data)
}

func Success(c echo.Context, msg string) error {
	return c.JSON(http.StatusOK, map[string]string{"success": msg})
}
