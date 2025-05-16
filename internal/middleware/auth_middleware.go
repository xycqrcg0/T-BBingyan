package middleware

import (
	"BBingyan/internal/config"
	"BBingyan/internal/controller/param"
	"BBingyan/internal/util"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func CheckJWT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			path := c.Request().URL.Path
			for _, skipPath := range config.Config.JWT.Skipper {
				if skipPath == path {
					return next(c)
				}
			}

			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, &param.Response{
					Status: false,
					Msg:    "",
				})
			}

			claims, err := util.ParseJWT(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, &param.Response{
					Status: false,
					Msg:    err.Error(),
				})
			}

			if ok := AdminUserFilter(path, claims.Permission); !ok {
				return c.JSON(http.StatusUnauthorized, &param.Response{
					Status: false,
					Msg:    "Not allowed",
				})
			}

			c.Set("identification", claims.Auth)
			c.Set("permission", claims.Permission)
			return next(c)
		}
	}
}

func AdminUserFilter(path string, per int) bool {
	if per == param.USER {
		for _, compared := range config.Config.JWT.Admin {
			parts := strings.Split(compared, "/")
			j := parts[len(parts)-1]
			if j[0] == '*' {
				if strings.HasPrefix(path, compared[:len(compared)-1]) {
					return false
				}
			} else if j[0] == ':' {
				if strings.HasPrefix(path, compared[:len(compared)-len(j)]) {
					return false
				}
			} else {
				if path == compared {
					return false
				}
			}
		}
	} else if per == param.ADMIN {
		for _, compared := range config.Config.JWT.User {
			parts := strings.Split(compared, "/")
			j := parts[len(parts)-1]
			if j[0] == '*' {
				if strings.HasPrefix(path, compared[:len(compared)-1]) {
					return false
				}
			} else if j[0] == ':' {
				if strings.HasPrefix(path, compared[:len(compared)-len(j)]) {
					return false
				}
			} else {
				if path == compared {
					return false
				}
			}
		}
	} else {
		return false
	}
	return true
}
