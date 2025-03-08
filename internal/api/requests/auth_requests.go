package requests

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	UserAgent string `json:"-"` // Заполняется из HTTP заголовка
	ClientIP  string `json:"-"` // Заполняется из IP адреса
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	UserAgent string `json:"-"` // Заполняется из HTTP заголовка
	ClientIP  string `json:"-"` // Заполняется из IP адреса
}

// RefreshTokenRequest запрос на обновление токенов
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh__token" validate:"required"`
	UserAgent    string `json:"-"` // Заполняется из HTTP заголовка
	ClientIP     string `json:"-"` // Заполняется из IP адреса
}

// LogoutRequest запрос на выход
type LogoutRequest struct {
	RefreshToken string `json:"refresh__token" validate:"required"`
}