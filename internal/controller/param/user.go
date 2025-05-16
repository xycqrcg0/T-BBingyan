package param

import "time"

type UserRequest struct {
	Code     string `json:"code"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserUpdateRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	Signature string `json:"signature"`
}

type UserLessInfoResponse struct {
	Email     string `json:"email"` //假设有跳转，还要靠这个字段，应该可以选择此字段前端不展示吧
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Likes     int    `json:"likes"`
	Follows   int    `json:"follows"`
}

type UserMoreInfoResponse struct {
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Signature  string    `json:"signature"`
	Likes      int       `json:"likes"`
	Follows    int       `json:"follows"`
	CreatedAt  time.Time `json:"created-at"`
	IsLiked    bool      `json:"is-liked"`
	IsFollowed bool      `json:"is-followed"`
}
