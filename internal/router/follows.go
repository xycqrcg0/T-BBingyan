package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func FollowsRouter(r *echo.Echo) {
	follows := r.Group("/follows")
	{
		follows.POST("/:email", controller.FollowUser)
		follows.DELETE("/:email", controller.UnFollowUser)
		follows.GET("/:email", controller.GetFollows)
		follows.GET("/fans", controller.GetFans) //这个目前是只能看自己的fans
	}
}
