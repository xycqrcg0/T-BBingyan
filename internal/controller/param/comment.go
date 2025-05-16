package param

import "time"

type CommentReq struct {
	Pid     int    `json:"pid"`
	Root    int    `json:"root"`
	Parent  int    `json:"parent"`
	Content string `json:"content"`
}

type CommentRes struct {
	Id        uint                 `json:"id"`
	Uid       string               `json:"uid"`
	Pid       int                  `json:"pid"`
	Root      int                  `json:"root"`
	Parent    int                  `json:"parent"`
	Content   string               `json:"content"`
	Replies   int                  `json:"replies"`
	Likes     int                  `json:"likes"`
	CreatedAt time.Time            `json:"created-at"`
	User      UserLessInfoResponse `json:"user"`
}
