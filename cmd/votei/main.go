package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Static("/assets", "../../web/assets")
	e.File("/", "../../web/static/index.html")
	e.File("/vote", "../../web/static/vote.html")
	e.File("/termos-uso-privacidade", "../../web/static/termos.html")
	e.Logger.Fatal(e.Start(":8080"))
}
