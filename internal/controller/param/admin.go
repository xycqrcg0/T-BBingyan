package param

type AdminReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type TagReq struct {
	Tag string `json:"tag"`
}
