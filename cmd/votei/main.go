package main

import (
	"log"
	"text/template"

	"github.com/Hanufu/votei/internal/config"
	"github.com/Hanufu/votei/internal/database"
	"github.com/Hanufu/votei/internal/router"
	"github.com/labstack/echo/v4"
)

func main() {
	// Inicializa o banco de dados
	var err error
	database.DB, err = database.InitDB()
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}
	defer func() {
		if err := database.DB.Close(); err != nil {
			log.Printf("Erro ao fechar o banco de dados: %v", err)
		}
	}()

	// Cria a tabela de votos, se não existir
	database.CreateVotesTable(database.DB)

	// Carrega as contagens de votos
	database.LoadVoteCounts(database.DB)

	// Carrega o template de resultados
	config.ResultTemplate, err = template.ParseFiles(config.StaticPath + config.ResultFile)
	if err != nil {
		log.Fatalf("Erro ao carregar o template de resultados: %v", err)
	}

	// Inicializa o Echo
	e := echo.New()

	// Configuração das rotas
	router.SetupRoutes(e)

	// Usando HTTPS com os certificados autoassinados (opcional)
	// err = e.StartTLS(":443", "cert.pem", "key.pem")
	// if err != nil {
	// 	log.Fatalf("Erro ao iniciar o servidor HTTPS: %v", err)
	// }

	// Inicia o servidor HTTP
	err = e.Start(":80")
	if err != nil {
		log.Fatalf("Erro ao iniciar o servidor HTTP: %v", err)
	}
}
