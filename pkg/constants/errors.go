package constants

import "errors"

// ErrNotFound ошибка, когда объект не найден
var ErrNotFound = errors.New("not found")

// Ошибки видео сервиса
var (
	ErrVideoNotFound     = errors.New("video not found")
	ErrVideoBlocked      = errors.New("video is blocked")
	ErrVideoPrivate      = errors.New("video is private")
	ErrInvalidStatus     = errors.New("invalid video status")
	ErrVideoProcessing   = errors.New("video is still processing")
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// Ошибки сервиса аутентификации
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)
