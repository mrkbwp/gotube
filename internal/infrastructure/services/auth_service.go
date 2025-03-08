package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"
	"time"

	"github.com/mrkbwp/gotube/internal/api/requests"
	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
	"github.com/mrkbwp/gotube/internal/domain/services"
	"github.com/mrkbwp/gotube/pkg/jwt"
)

// AuthService реализует интерфейс AuthService
type AuthService struct {
	userRepo        repositories.UserRepository
	tokenRepo       repositories.TokenRepository
	passwordService *jwt.PasswordService
	jwtService      *jwt.JWTService
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	passwordService *jwt.PasswordService,
	jwtService *jwt.JWTService,
) services.AuthService {
	return &AuthService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req requests.RegisterRequest) (*entity.User, *jwt.TokenPair, error) {
	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, nil, constants.ErrUserAlreadyExists
	}

	// Хешируем пароль
	passwordHash, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем нового пользователя
	user := &entity.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         constants.RoleUser,
	}

	// Сохраняем пользователя в БД
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токены
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Сохраняем refresh токен в БД
	refreshToken := &entity.Token{
		UserID:       user.ID,
		RefreshToken: tokenPair.RefreshToken,
		UserAgent:    req.UserAgent,
		ClientIP:     req.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(s.jwtService.GetRefreshTokenTTL()),
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return user, tokenPair, nil
}

// Login авторизует пользователя
func (s *AuthService) Login(ctx context.Context, req requests.LoginRequest) (*entity.User, *jwt.TokenPair, error) {
	// Ищем пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, nil, constants.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем пароль
	if !s.passwordService.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, nil, constants.ErrInvalidCredentials
	}

	// Генерируем токены
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Сохраняем refresh токен в БД
	refreshToken := &entity.Token{
		UserID:       user.ID,
		RefreshToken: tokenPair.RefreshToken,
		UserAgent:    req.UserAgent,
		ClientIP:     req.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(s.jwtService.GetRefreshTokenTTL()),
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return user, tokenPair, nil
}

// RefreshTokens обновляет пару токенов
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string, userAgent string, clientIP string) (*jwt.TokenPair, error) {
	// Проверяем refresh токен
	userID, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Получаем токен из БД
	token, err := s.tokenRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Проверяем, что токен не заблокирован
	if token.IsBlocked {
		return nil, constants.ErrInvalidToken
	}

	// Проверяем, что токен не истек
	if token.ExpiresAt.Before(time.Now()) {
		return nil, constants.ErrTokenExpired
	}

	tokenUserID := &token.UserID

	// Проверяем соответствие UserID в токене и в БД
	if userID != tokenUserID {
		return nil, constants.ErrInvalidToken
	}

	// Блокируем старый токен
	if err := s.tokenRepo.UpdateBlockStatus(ctx, token.ID, true); err != nil {
		return nil, fmt.Errorf("failed to block old token: %w", err)
	}

	// Генерируем новую пару токенов
	tokenPair, err := s.jwtService.GenerateTokenPair(token.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Сохраняем новый refresh токен в БД
	newToken := &entity.Token{
		UserID:       token.UserID,
		RefreshToken: tokenPair.RefreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(s.jwtService.GetRefreshTokenTTL()),
	}

	if err := s.tokenRepo.Create(ctx, newToken); err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return tokenPair, nil
}

// Logout выход из системы
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// Получаем токен из БД
	token, err := s.tokenRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil // Если токен не найден, считаем операцию успешной
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Блокируем токен
	return s.tokenRepo.UpdateBlockStatus(ctx, token.ID, true)
}
