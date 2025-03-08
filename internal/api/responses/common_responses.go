package responses

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// SuccessResponse представляет успешный ответ
type SuccessResponse struct {
	Success bool `json:"success"`
}

// PaginatedResponse представляет ответ с пагинацией
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int64       `json:"total"`
}