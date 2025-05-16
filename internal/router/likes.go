package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func LikesRouter(r *echo.Echo) {
	likes := r.Group("/likes")
	{
		likes.POST("/user/:email", controller.LikeUser)
		likes.POST("/post/:id", controller.LikePost)
		likes.POST("/comment/:id", controller.LikeComment)
	}
}
