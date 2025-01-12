package storage

import (
	"database/sql"
	"fmt"
	"shortener/internal/config"
	"sync"
)

type StorageDB struct {
	urlStorage map[string]string
	mu         sync.Mutex
	dbConn     *sql.DB
}

func NewStorageDB(c *config.Config) *StorageDB {
	dbConn, err := sql.Open("pgx", c.DBConnection)
	if err != nil {
		_ = fmt.Errorf("unable open database: %v", err)
		return nil
	}

	return &StorageDB{
		urlStorage: make(map[string]string),
		dbConn:     dbConn,
	}
}

func (s *StorageDB) UpdateData(shortID, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.dbConn.Begin()
	if err != nil {
		fmt.Printf("Error beginning transaction: %v", err)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Printf("Error transaction rollback: %v\n", err)
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO shortener_db (short_url, full_url) VALUES ($1, $2)")
	if err != nil {
		fmt.Printf("Error preparing statement: %v\n", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(shortID, originalURL)
	if err != nil {
		fmt.Printf("Error inserting row to DB: %v\n", err)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
	}
}

func (s *StorageDB) GetData(shortID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var originalURL string
	err := s.dbConn.QueryRow("SELECT full_url FROM shortener_db WHERE short_url=$1", shortID).Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (s *StorageDB) Ping() error {
	return s.dbConn.Ping()
}

func CreateTable(s *StorageDB) error {
	query := `
    CREATE TABLE IF NOT EXISTS shortener_db (
        id SERIAL PRIMARY KEY,
        short_url TEXT UNIQUE NOT NULL,
        full_url TEXT NOT NULL
    );`
	_, err := s.dbConn.Exec(query)
	return err
}
