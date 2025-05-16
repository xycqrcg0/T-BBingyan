package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func CommentRouter(r *echo.Echo) {
	comment := r.Group("/comment")
	{
		comment.POST("/new", controller.CreateComment)
		comment.DELETE("/del/:id", controller.DeleteComment)
		comment.GET("/comments", controller.GetComments)
		comment.GET("/replies", controller.GetReplies)
	}
}
