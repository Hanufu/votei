package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	DB         *sql.DB
	DBLock     sync.Mutex
	VoteCounts = struct {
		sync.RWMutex
		Counts map[int]int
	}{Counts: make(map[int]int)}
)

// Função para inicializar o banco de dados
func InitDB() (*sql.DB, error) {
	// Pega as credenciais do banco de dados a partir das variáveis de ambiente
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Verifica se as variáveis de ambiente estão definidas
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		return nil, fmt.Errorf("uma ou mais variáveis de ambiente não estão definidas")
	}

	// Constrói a string de conexão utilizando as variáveis de ambiente
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable port=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort)

	// Abre a conexão com o banco de dados
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir a conexão com o banco de dados: %w", err)
	}

	// Verifica se a conexão foi bem-sucedida
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco de dados PostgreSQL: %w", err)
	}

	fmt.Println("Conectado ao banco de dados PostgreSQL!")
	return db, nil
}

// Função para criar a tabela de votos
func CreateVotesTable(db *sql.DB) {
	createVotesTable := `CREATE TABLE IF NOT EXISTS votes (
        id SERIAL PRIMARY KEY,
        ip_address TEXT,
        user_agent TEXT,
        cookie_id TEXT,
        timestamp TIMESTAMPTZ,
        referer TEXT,
        language TEXT,
        browser TEXT,
        candidate_number INTEGER,
        latitude TEXT,
        longitude TEXT
    );`

	_, err := db.Exec(createVotesTable)
	if err != nil {
		fmt.Println("Erro ao criar tabela de votos:", err)
	} else {
		fmt.Println("Tabela de votos criada ou já existe.")
	}
}

// Função para carregar contagens de votos
func LoadVoteCounts(db *sql.DB) {
	rows, err := db.Query("SELECT candidate_number, COUNT(*) FROM votes GROUP BY candidate_number")
	if err != nil {
		fmt.Println("Erro ao carregar contagens de votos:", err)
		return
	}
	defer rows.Close()

	VoteCounts.Lock()
	defer VoteCounts.Unlock()

	for rows.Next() {
		var candidateNumber, count int
		if err := rows.Scan(&candidateNumber, &count); err != nil {
			fmt.Println("Erro ao escanear contagens de votos:", err)
			return
		}
		VoteCounts.Counts[candidateNumber] = count
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Erro ao iterar sobre as linhas:", err)
		return
	}

	fmt.Println("Contagens de votos carregadas com sucesso.")
}
