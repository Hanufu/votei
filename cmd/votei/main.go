package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

// Constantes e variáveis
const (
	staticPath = "../../web/static/"
	indexFile  = "index.html"
	voteFile   = "vote.html"
	termosFile = "termos.html"
	dbFile     = "votes.db"
)

var (
	db *sql.DB
)

// Candidate estrutura para candidatos
type Candidate struct {
	Number      int
	VoteNumbers int64
}

// Vote estrutura para armazenar informações do voto
type Vote struct {
	IPAddress string
	UserAgent string
	CookieID  string
	Timestamp time.Time
	Referer   string
	Language  string
	Browser   string
}

// Função principal
func main() {
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("Erro ao abrir o banco de dados:", err)
		return
	}
	defer db.Close()

	// Criação das tabelas se não existirem
	createTables()

	e := echo.New()

	e.Static("/assets", "/../../web/assets")

	e.GET("/", serveFile(indexFile))
	e.GET("/vote", serveFile(voteFile))
	e.GET("/termos-uso-privacidade", serveFile(termosFile))
	e.POST("/vote", voteHandler)
	e.GET("/result", handleResult)

	e.Logger.Fatal(e.Start(":8080"))
}

// Criação das tabelas se não existirem
func createTables() {
	createCandidatesTable := `CREATE TABLE IF NOT EXISTS candidates (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        number INTEGER NOT NULL,
        vote_numbers INTEGER DEFAULT 0
    );`
	createVotesTable := `CREATE TABLE IF NOT EXISTS votes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address TEXT,
        user_agent TEXT,
        cookie_id TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        referer TEXT,
        language TEXT,
        browser TEXT
    );`
	_, err := db.Exec(createCandidatesTable)
	if err != nil {
		fmt.Println("Erro ao criar a tabela de candidatos:", err)
	}
	_, err = db.Exec(createVotesTable)
	if err != nil {
		fmt.Println("Erro ao criar a tabela de votos:", err)
	}
}

// Handler para servir arquivos estáticos
func serveFile(fileName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.File(staticPath + fileName)
	}
}

// Gera um identificador único baseado em IP e User-Agent
func getUniqueIdentifier(c echo.Context) string {
	ip := c.Request().Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.RealIP()
	}
	userAgent := c.Request().Header.Get("User-Agent")
	return ip + "-" + userAgent
}

// Gera e recupera o ID do cookie do usuário
func generateCookieID(c echo.Context) string {
	cookie, err := c.Cookie("voter_id")
	if err != nil {
		if err == http.ErrNoCookie {
			newID := uuid.New().String()
			c.SetCookie(&http.Cookie{
				Name:    "voter_id",
				Value:   newID,
				Path:    "/",
				Expires: time.Now().Add(24 * time.Hour), // Expira em 24 horas
			})
			return newID
		}
		return ""
	}
	return cookie.Value
}

// Verifica se o identificador ou cookie já votou
func hasVoted(identifier string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM votes WHERE ip_address = ? OR cookie_id = ?", identifier, identifier).Scan(&count)
	if err != nil {
		fmt.Println("Erro ao verificar votos:", err)
		return false
	}
	return count > 0
}

// Registra o voto com base no identificador e no cookie
func registerVote(vote Vote) {
	_, err := db.Exec("INSERT INTO votes (ip_address, user_agent, cookie_id, timestamp, referer, language, browser) VALUES (?, ?, ?, ?, ?, ?, ?)",
		vote.IPAddress, vote.UserAgent, vote.CookieID, vote.Timestamp, vote.Referer, vote.Language, vote.Browser)
	if err != nil {
		fmt.Println("Erro ao registrar voto:", err)
	}
}

// Atualiza o número de votos do candidato
func updateVoteCount(candidateNumber int) {
	_, err := db.Exec("UPDATE candidates SET vote_numbers = vote_numbers + 1 WHERE number = ?", candidateNumber)
	if err != nil {
		fmt.Println("Erro ao atualizar contagem de votos:", err)
	}
}

// Imprime o log com informações detalhadas
func logVote(vote Vote) {
	fmt.Printf("Voto registrado:\n")
	fmt.Printf("IP: %s\n", vote.IPAddress)
	fmt.Printf("User-Agent: %s\n", vote.UserAgent)
	fmt.Printf("Cookie ID: %s\n", vote.CookieID)
	fmt.Printf("Timestamp: %s\n", vote.Timestamp.Format(time.RFC3339))
	fmt.Printf("Referer: %s\n", vote.Referer)
	fmt.Printf("Language: %s\n", vote.Language)
	fmt.Printf("Browser: %s\n", vote.Browser)
}

// Handler de votação
// Handler de votação
func voteHandler(c echo.Context) error {
	identifier := getUniqueIdentifier(c)
	cookieID := generateCookieID(c)
	ip := c.RealIP()
	userAgent := c.Request().Header.Get("User-Agent")
	referer := c.Request().Header.Get("Referer")
	acceptLanguage := c.Request().Header.Get("Accept-Language")

	// Determina o navegador (simplificação, pode ser aprimorada)
	var browser string
	if userAgent != "" {
		if strings.Contains(userAgent, "Firefox") {
			browser = "Firefox"
		} else if strings.Contains(userAgent, "Chrome") {
			browser = "Chrome"
		} else if strings.Contains(userAgent, "Safari") {
			browser = "Safari"
		} else {
			browser = "Other"
		}
	}

	// Cria a estrutura de voto
	vote := Vote{
		IPAddress: ip,
		UserAgent: userAgent,
		CookieID:  cookieID,
		Timestamp: time.Now(),
		Referer:   referer,
		Language:  acceptLanguage,
		Browser:   browser,
	}

	// Verifica se o identificador ou o cookie já votou
	if hasVoted(identifier) || hasVoted(cookieID) {
		return c.JSON(http.StatusForbidden, "Você já votou.")
	}

	// Recebe o número do candidato do input
	candidateNumber := c.FormValue("candidate_number")

	// Verifica se o número é "00" para voto em branco
	var candidateNumberInt int
	if candidateNumber == "00" {
		candidateNumberInt = 0 // Número 0 para voto em branco
	} else {
		var err error
		candidateNumberInt, err = strconv.Atoi(candidateNumber)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Número do candidato inválido.")
		}
	}

	// Registra o voto
	registerVote(vote)

	// Atualiza a contagem de votos do respectivo candidato
	updateVoteCount(candidateNumberInt)

	// Imprime o log com informações detalhadas
	logVote(vote)

	return c.String(http.StatusOK, "Voto registrado com sucesso.")
}

// Handler para retornar os resultados dos candidatos
func handleResult(c echo.Context) error {
	rows, err := db.Query("SELECT number, vote_numbers FROM candidates")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Erro ao consultar resultados")
	}
	defer rows.Close()

	results := make(map[int]int64)
	for rows.Next() {
		var number int
		var voteNumbers int64
		if err := rows.Scan(&number, &voteNumbers); err != nil {
			return c.JSON(http.StatusInternalServerError, "Erro ao processar resultados")
		}
		results[number] = voteNumbers
	}

	return c.JSON(http.StatusOK, results)
}
