package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func PostRouter(r *echo.Echo) {
	post := r.Group("/posts")
	{
		post.POST("/new", controller.AddPost)
		post.DELETE("/del/:id", controller.DeletePost)
		post.GET("/:email", controller.GetPostByEmail)
		post.GET("", controller.GetPostByTag)
		post.POST("/search", controller.SearchPost)
		post.GET("/search/:id", controller.GetPostById)
	}
}
