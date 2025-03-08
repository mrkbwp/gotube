package jwt

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordService предоставляет функции для работы с паролями
type PasswordService struct {
	cost int
}

// NewPasswordService создает новый PasswordService
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: bcrypt.DefaultCost,
	}
}

// HashPassword создает хеш пароля
func (s *PasswordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash проверяет пароль по хешу
func (s *PasswordService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
