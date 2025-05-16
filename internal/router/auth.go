package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func AuthRouter(r *echo.Echo) {
	auth := r.Group("/auth")
	{
		auth.POST("/register/code", controller.RegisterForCode)
		auth.POST("/register", controller.Register)
		auth.POST("/login/code", controller.LoginForCode)
		auth.POST("/login/:style", controller.Login)
	}
}
