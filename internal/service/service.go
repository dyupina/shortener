package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/9ssi7/nanoid"
)

var ErrDuplicateURL = errors.New("duplicate URL")

type Service interface {
	GetShortURL_db(originalURL string) (string, error)
}
type Serv struct {
}

const insertRow = `
INSERT INTO urls (short_url, full_url) VALUES ($1, $2)
ON CONFLICT (full_url) DO NOTHING
RETURNING short_url`

func (s *Serv) GetShortURLDB(originalURL string, db *sql.DB) (string, error) {
	var shortURL string
	var retErr error
	shortID := GenerateShortID()

	row := db.QueryRow(insertRow, shortID, originalURL)
	row.Scan(&shortURL)
	retErr = nil

	if shortURL == "" {
		// Получение существующего сокращенного URL
		row := db.QueryRow(
			"SELECT short_url FROM urls WHERE full_url = $1", originalURL)
		err := row.Scan(&shortURL)
		if err != nil {
			return "", fmt.Errorf("error select query: %v", err)
		}
		retErr = ErrDuplicateURL
		fmt.Println("Existing short URL:", shortURL)
	}

	return shortURL, retErr
}

func GenerateShortID() string {
	id, _ := nanoid.New()
	return id
}
