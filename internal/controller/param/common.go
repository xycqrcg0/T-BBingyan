package param

var (
	VALID   = "1"
	INVALID = "0"

	TIME      = "0"
	TIMEDESC  = "1"
	REPLY     = "2"
	REPLYDESC = "3"

	USER  = 0
	ADMIN = 1
)

type Response struct {
	Status bool        `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}
