package param

type FollowUser struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Likes     int    `json:"likes"`
	Follows   int    `json:"follows"`
}

type FollowsResponse struct {
	Page     int          `json:"page"`
	PageSize int          `json:"page-size"`
	Follows  []FollowUser `json:"follows"`
}
