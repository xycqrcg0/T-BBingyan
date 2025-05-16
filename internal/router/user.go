package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func UserRouter(r *echo.Echo) {
	user := r.Group("/user")
	{
		user.GET("/info/:email", controller.GetUSerInfo)
	}
}
