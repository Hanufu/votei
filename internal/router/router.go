package router

import (
	"github.com/Hanufu/votei/internal/config"
	"github.com/Hanufu/votei/internal/handlers"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	e.Static("/assets", config.StaticPath)
	e.GET("/", handlers.ServeFile(config.IndexFile))
	e.GET("/vote", handlers.ServeFile(config.VoteFile))
	e.GET("/termos-uso-privacidade", handlers.ServeFile(config.TermosFile))
	e.POST("/vote", handlers.VoteHandler) // Certifique-se de que VoteHandler é uma função
	e.GET("/result", handlers.ResultHandler)
}
