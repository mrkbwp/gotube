package middleware

import (
	"github.com/google/uuid"
	"github.com/mrkbwp/gotube/pkg/jwt"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware создает middleware для проверки JWT токена
func AuthMiddleware(jwtService *jwt.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Получаем заголовок Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "Missing authorization header")
			}

			// Проверяем формат (Bearer <token>)
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(401, "Invalid authorization format")
			}

			// Получаем токен
			token := parts[1]

			// Проверяем токен
			userID, err := jwtService.ValidateAccessToken(token)
			if err != nil {
				return echo.NewHTTPError(401, "Invalid token: "+err.Error())
			}

			// Сохраняем ID пользователя в контексте
			c.Set("userID", uuid.MustParse(userID))

			return next(c)
		}
	}
}
