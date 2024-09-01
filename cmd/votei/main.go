package main

import (
	"fmt"
	"text/template"

	"github.com/Hanufu/votei/internal/config"
	"github.com/Hanufu/votei/internal/database"
	"github.com/Hanufu/votei/internal/router"
	"github.com/labstack/echo/v4"
)

func main() {
	var err error
	database.DB, err = database.InitDB() // Inicializa o banco de dados
	if err != nil {
		fmt.Println("Erro ao abrir o banco de dados:", err)
		return
	}
	defer database.DB.Close()

	database.CreateVotesTable()
	database.LoadVoteCounts() // Carregue a contagem de votos após criar a tabela

	// Carrega o template de resultados
	config.ResultTemplate, err = template.ParseFiles(config.StaticPath + config.ResultFile)
	if err != nil {
		fmt.Println("Erro ao carregar o template de resultados:", err)
		return
	}

	e := echo.New()

	// Configuração das rotas
	router.SetupRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
