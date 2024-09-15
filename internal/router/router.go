package router

import (
	"github.com/Hanufu/votei/internal/config"
	"github.com/Hanufu/votei/internal/handlers"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	e.Static("/assets", config.StaticPath)
	e.GET("/", handlers.ServeFile(config.InfoFile))
	e.POST("/", handlers.GetEmailHandler)
	//e.GET("/vote", handlers.ServeFile(config.VoteFile))
	e.GET("/termos-uso-privacidade", handlers.ServeFile(config.TermosFile))
	//e.POST("/vote", handlers.VoteHandler)
	//e.GET("/result", handlers.ResultHandler)
	e.GET("/admin", handlers.ServeFile(config.AdminLogin))
	e.POST("/admin", handlers.AdminLoginHandler)

	e.GET("/download/:filename", handlers.DownloadFileHandler)

}
