package router

import (
	"github.com/labstack/echo/v4"
)

func SetupRouter(r *echo.Echo) {
	AuthRouter(r)
	UserRouter(r)
	FollowsRouter(r)
	PostRouter(r)
	LikesRouter(r)
	CommentRouter(r)
	AdminRouter(r)
}
