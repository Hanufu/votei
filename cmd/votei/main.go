package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"
)

const (
	staticPath = "../../web/static/"
	indexFile  = "index.html"
	voteFile   = "vote.html"
	termosFile = "termos.html"
	resultFile = "result.html"
	dbFile     = "../../database/votes.db"
)

var (
	db     *sql.DB
	dbLock sync.Mutex
)

type Vote struct {
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	CookieID        string    `json:"cookie_id"`
	Timestamp       time.Time `json:"timestamp"`
	Referer         string    `json:"referer"`
	Language        string    `json:"language"`
	Browser         string    `json:"browser"`
	CandidateNumber int       `json:"candidate_number"`
	Latitude        string    `json:"latitude"`
	Longitude       string    `json:"longitude"`
}

var voteCounts = struct {
	sync.RWMutex
	counts map[int]int
}{
	counts: make(map[int]int),
}

func main() {
	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Println("Erro ao abrir o banco de dados:", err)
		return
	}
	defer db.Close()

	createVotesTable()
	loadVoteCounts() // Carregue a contagem de votos após criar a tabela

	e := echo.New()

	e.Static("/assets", staticPath+"assets")
	e.GET("/", serveFile(indexFile))
	e.GET("/vote", serveFile(voteFile))
	e.GET("/termos-uso-privacidade", serveFile(termosFile))
	e.POST("/vote", voteHandler)
	e.GET("/result", resultHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

func createVotesTable() {
	createVotesTable := `CREATE TABLE IF NOT EXISTS votes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address TEXT,
        user_agent TEXT,
        cookie_id TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        referer TEXT,
        language TEXT,
        browser TEXT,
        candidate_number INTEGER,
        latitude TEXT,
        longitude TEXT
    );`
	if _, err := db.Exec(createVotesTable); err != nil {
		fmt.Println("Erro ao criar a tabela de votos:", err)
	}
}

func serveFile(fileName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.File(staticPath + fileName)
	}
}

func getUniqueIdentifier(c echo.Context) string {
	ip := c.Request().Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.RealIP()
	}
	userAgent := c.Request().Header.Get("User-Agent")
	return ip + "-" + userAgent
}

func generateCookieID(c echo.Context) string {
	cookie, err := c.Cookie("voter_id")
	if err != nil {
		if err == http.ErrNoCookie {
			newID := uuid.New().String()
			c.SetCookie(&http.Cookie{
				Name:    "voter_id",
				Value:   newID,
				Path:    "/",
				Expires: time.Now().Add(24 * time.Hour),
			})
			return newID
		}
		return ""
	}
	return cookie.Value
}

func hasVoted(identifier string) bool {
	var count int
	// Verifica se o identificador já está presente na tabela de votos
	if err := db.QueryRow("SELECT COUNT(*) FROM votes WHERE ip_address = ? OR cookie_id = ?", identifier, identifier).Scan(&count); err != nil {
		fmt.Println("Erro ao verificar votos:", err)
		return false
	}
	return count > 0
}
func registerVote(vote Vote) {
	dbLock.Lock()
	defer dbLock.Unlock()

	// Insere o voto no banco de dados
	_, err := db.Exec("INSERT INTO votes (ip_address, user_agent, cookie_id, timestamp, referer, language, browser, candidate_number, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		vote.IPAddress, vote.UserAgent, vote.CookieID, vote.Timestamp, vote.Referer, vote.Language, vote.Browser, vote.CandidateNumber, vote.Latitude, vote.Longitude)
	if err != nil {
		fmt.Println("Erro ao registrar voto:", err)
		return
	}

	// Atualiza a contagem de votos em memória
	voteCounts.Lock()
	defer voteCounts.Unlock()
	voteCounts.counts[vote.CandidateNumber]++
}

func logVote(vote Vote) {
	fmt.Printf("Voto registrado:\n")
	fmt.Printf("IP: %s\n", vote.IPAddress)
	fmt.Printf("User-Agent: %s\n", vote.UserAgent)
	fmt.Printf("Cookie ID: %s\n", vote.CookieID)
	fmt.Printf("Timestamp: %s\n", vote.Timestamp.Format(time.RFC3339))
	fmt.Printf("Referer: %s\n", vote.Referer)
	fmt.Printf("Language: %s\n", vote.Language)
	fmt.Printf("Browser: %s\n", vote.Browser)
	fmt.Printf("Número do Candidato: %d\n", vote.CandidateNumber)
	fmt.Printf("Latitude: %s\n", vote.Latitude)
	fmt.Printf("Longitude: %s\n", vote.Longitude)
}

func resultHandler(c echo.Context) error {
	message := c.QueryParam("message")

	voteCounts.RLock()
	defer voteCounts.RUnlock()

	// Monta a página HTML com a contagem de votos e a mensagem, se houver
	html := `<html>
        <head><title>Resultado dos Votos</title></head>
        <body>
            <h1>Resultado dos Votos</h1>`

	if message != "" {
		html += `<p><strong>` + message + `</strong></p>`
	}

	html += `<p>Votos em branco: ` + strconv.Itoa(voteCounts.counts[0]) + `</p>
            <p>Votos no 45: ` + strconv.Itoa(voteCounts.counts[45]) + `</p>
            <p>Votos no 13: ` + strconv.Itoa(voteCounts.counts[13]) + `</p>
        </body>
    </html>`

	return c.HTML(http.StatusOK, html)
}

func voteHandler(c echo.Context) error {
	identifier := getUniqueIdentifier(c)
	cookieID := generateCookieID(c)
	ip := c.RealIP()
	userAgent := c.Request().Header.Get("User-Agent")
	referer := c.Request().Header.Get("Referer")
	acceptLanguage := c.Request().Header.Get("Accept-Language")

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

	candidateNumber := c.FormValue("candidate_number")
	latitude := c.FormValue("latitude")
	longitude := c.FormValue("longitude")

	var candidateNumberInt int
	if candidateNumber == "00" {
		candidateNumberInt = 0
	} else {
		var err error
		candidateNumberInt, err = strconv.Atoi(candidateNumber)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Número do candidato inválido."})
		}
	}

	// Verifique se o usuário já votou
	if hasVoted(identifier) || hasVoted(cookieID) {
		// Redirecione para a página de resultados com uma mensagem
		return c.Redirect(http.StatusSeeOther, "/result?message=Você já votou antes! Seu voto já foi registrado e não será computado novamente.")
	}

	// Registre o voto se ainda não foi registrado
	vote := Vote{
		IPAddress:       ip,
		UserAgent:       userAgent,
		CookieID:        cookieID,
		Timestamp:       time.Now(),
		Referer:         referer,
		Language:        acceptLanguage,
		Browser:         browser,
		CandidateNumber: candidateNumberInt,
		Latitude:        latitude,
		Longitude:       longitude,
	}

	registerVote(vote)
	logVote(vote)

	return c.Redirect(http.StatusSeeOther, "/result")
}

func handleResult(c echo.Context) error {
	rows, err := db.Query("SELECT candidate_number, COUNT(*) as vote_count FROM votes GROUP BY candidate_number ORDER BY vote_count DESC")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao obter resultados."})
	}
	defer rows.Close()

	type Result struct {
		CandidateNumber int   `json:"candidate_number"`
		VoteCount       int64 `json:"vote_count"`
	}

	var results []Result
	for rows.Next() {
		var result Result
		if err := rows.Scan(&result.CandidateNumber, &result.VoteCount); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao processar resultados."})
		}
		results = append(results, result)
	}

	return c.JSON(http.StatusOK, results)
}

func loadVoteCounts() {
	voteCounts.Lock()
	defer voteCounts.Unlock()

	rows, err := db.Query("SELECT candidate_number, COUNT(*) as vote_count FROM votes GROUP BY candidate_number")
	if err != nil {
		fmt.Println("Erro ao carregar a contagem de votos:", err)
		return
	}
	defer rows.Close()

	// Limpe a contagem de votos atual
	voteCounts.counts = make(map[int]int)

	for rows.Next() {
		var candidateNumber int
		var voteCount int64
		if err := rows.Scan(&candidateNumber, &voteCount); err != nil {
			fmt.Println("Erro ao processar resultados:", err)
			return
		}
		voteCounts.counts[candidateNumber] = int(voteCount)
	}
}
