package router

import (
	"BBingyan/internal/controller"
	"github.com/labstack/echo/v4"
)

func AdminRouter(r *echo.Echo) {
	admin := r.Group("/admin")
	{
		admin.POST("/login", controller.AdminLogin)
		admin.POST("/new", controller.CreateAdmin)
		admin.POST("/del", controller.DeleteAdmin)
		admin.POST("/tag", controller.CreateTag)
		admin.POST("/post/del", controller.AdminDeletePost)
		admin.POST("/comment/del", controller.AdminDeleteComment)
	}
}
