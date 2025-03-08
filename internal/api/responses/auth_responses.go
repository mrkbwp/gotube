package responses

import (
	"github.com/mrkbwp/gotube/pkg/jwt"
	"time"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// TokenResponse представляет ответ с токенами
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// AuthResponse представляет ответ при аутентификации
type AuthResponse struct {
	User  *entity.User   `json:"user"`
	Token *TokenResponse `json:"token"`
}

// NewTokenResponse создает ответ с токенами
func NewTokenResponse(tokenPair *jwt.TokenPair) *TokenResponse {
	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    "bearer",
	}
}

// NewAuthResponse создает ответ при аутентификации
func NewAuthResponse(user *entity.User, tokenPair *jwt.TokenPair) *AuthResponse {
	return &AuthResponse{
		User:  user,
		Token: NewTokenResponse(tokenPair),
	}
}
