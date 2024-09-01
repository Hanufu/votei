package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	DB         *sql.DB
	DBLock     sync.Mutex
	VoteCounts = struct {
		sync.RWMutex
		Counts map[int]int
	}{Counts: make(map[int]int)}
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "../../database/votes.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateVotesTable() {
	createVotesTable := `CREATE TABLE IF NOT EXISTS votes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT,
		user_agent TEXT,
		cookie_id TEXT,
		timestamp DATETIME,
		referer TEXT,
		language TEXT,
		browser TEXT,
		candidate_number INTEGER,
		latitude TEXT,
		longitude TEXT
	);`

	_, err := DB.Exec(createVotesTable)
	if err != nil {
		fmt.Println("Erro ao criar tabela de votos:", err)
	}
}

func LoadVoteCounts() {
	rows, err := DB.Query("SELECT candidate_number, COUNT(*) FROM votes GROUP BY candidate_number")
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
}
