package main

import (
	"BBingyan/internal/config"
	"BBingyan/internal/log"
	middleware2 "BBingyan/internal/middleware"
	"BBingyan/internal/model"
	"BBingyan/internal/router"
	"BBingyan/internal/util"
	"github.com/labstack/echo/v4"
	"time"
)

func main() {
	config.InitConfig()
	model.Init()
	r := echo.New()
	r.Use(middleware2.CheckJWT())
	router.SetupRouter(r)

	go func() {
		ticker := time.Tick(time.Hour * 4)
		for {
			select {
			case <-ticker:
				if err := util.Archive(); err != nil {
					log.Errorf("Fail to archive information")
				}
			}
		}
	}()

	r.Logger.Fatal(r.Start(config.Config.Port))
}
