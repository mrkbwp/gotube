package requests

type CommentRequest struct {
	Text     string  `json:"text"`
	ParentID *string `json:"parent_id,omitempty"`
}
