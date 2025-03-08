package dto

import "github.com/mrkbwp/gotube/internal/domain/entity"

type VideoUserResponse struct {
	entity.Video
	Liked    bool `json:"liked"`
	Disliked bool `json:"disliked"`
}
