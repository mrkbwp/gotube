package requests

// UpdateVideoRequest запрос на обновление информации о видео
type UpdateVideoRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=100"`
	Description string `json:"description"`
	CategoryID  string `json:"category_id" validate:"required,uuid"`
}

// VideoIDRequest запрос с ID видео
type VideoIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}