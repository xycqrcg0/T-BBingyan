package param

import "time"

type PostRequest struct {
	Title   string `json:"title"`
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type PostResponse struct {
	ID        uint                 `json:"id"`
	Title     string               `json:"title"`
	Tag       string               `json:"tag"`
	Content   string               `json:"content"`
	Likes     int                  `json:"likes"`
	Replies   int                  `json:"replies"`
	CreatedAt time.Time            `json:"created-at"`
	User      UserLessInfoResponse `json:"user"`
}
