package handlers

import (
	"github.com/mrkbwp/gotube/pkg/constants"
	"github.com/mrkbwp/gotube/pkg/validator"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkbwp/gotube/internal/api/requests"
	"github.com/mrkbwp/gotube/internal/api/responses"
	"github.com/mrkbwp/gotube/internal/domain/services"
)

// AuthHandler обработчик аутентификации
type AuthHandler struct {
	authService services.AuthService
	validator   *validator.Validator
}

// NewAuthHandler создает новый обработчик
func NewAuthHandler(authService services.AuthService, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

// Register обработка регистрации
func (h *AuthHandler) Register(c echo.Context) error {
	var req requests.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Ошибка в данных запроса")
	}

	// Установка UserAgent и IP через контекст Echo
	req.UserAgent = c.Request().UserAgent()
	req.ClientIP = c.RealIP()

	user, tokens, err := h.authService.Register(c.Request().Context(), req)
	if err != nil {
		switch {
		case err == constants.ErrUserAlreadyExists:
			return responses.Error(c, http.StatusConflict, "Пользователь уже существует")
		default:
			return responses.Error(c, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
	}

	return responses.JSON(c, http.StatusCreated, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// Login обработка входа
func (h *AuthHandler) Login(c echo.Context) error {
	var req requests.LoginRequest
	if err := c.Bind(&req); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Ошибка в данных запроса")
	}

	req.UserAgent = c.Request().UserAgent()
	req.ClientIP = c.RealIP()

	user, tokens, err := h.authService.Login(c.Request().Context(), req)
	if err != nil {
		switch {
		case err == constants.ErrInvalidCredentials:
			return responses.Error(c, http.StatusUnauthorized, "Неверный email или пароль")
		default:
			return responses.Error(c, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
	}

	return responses.JSON(c, http.StatusOK, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// Refresh обновление токенов
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req requests.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Ошибка в данных запроса")
	}

	userAgent := c.Request().UserAgent()
	clientIP := c.RealIP()

	tokens, err := h.authService.RefreshTokens(c.Request().Context(), req.RefreshToken, userAgent, clientIP)
	if err != nil {
		switch {
		case err == constants.ErrInvalidToken:
			return responses.Error(c, http.StatusUnauthorized, "Неверный refresh token")
		case err == constants.ErrTokenExpired:
			return responses.Error(c, http.StatusUnauthorized, "Refresh token истек")
		default:
			return responses.Error(c, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
	}

	return responses.JSON(c, http.StatusOK, tokens)
}

// Logout обработка выхода
func (h *AuthHandler) Logout(c echo.Context) error {
	var req requests.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Ошибка в данных запроса")
	}

	err := h.authService.Logout(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Ошибка при выходе")
	}

	return responses.Success(c, "Вы успешно вышли")
}
