package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/9ssi7/nanoid"
)

// ErrDuplicateURL - error when the original URL already exists in the system.
var ErrDuplicateURL = errors.New("duplicate URL")

// Repository - interface for working with shortened URLs.
type Repository interface {
	GetShortURL_db(originalURL string) (string, error)
}

// Repo - structure for interacting with the data storage.
type Repo struct {
}

const insertRow = `
INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3)
ON CONFLICT (original_url) DO NOTHING
RETURNING short_url`

// GetShortURLDB returns the shortened URL for the given original URL and user ID.
// If the URL already exists, it returns the existing shortened URL with the error ErrDuplicateURL.
func (s *Repo) GetShortURLDB(userID, originalURL string, db *sql.DB) (string, error) {
	var shortURL string
	var retErr error
	shortID := GenerateShortID()

	row := db.QueryRow(insertRow, userID, shortID, originalURL)
	_ = row.Scan(&shortURL)

	retErr = nil

	if shortURL == "" {
		// Получение существующего сокращенного URL
		row := db.QueryRow(
			"SELECT short_url FROM urls WHERE original_url = $1", originalURL)
		err := row.Scan(&shortURL)
		if err != nil {
			return "", fmt.Errorf("error select query: %v", err)
		}
		retErr = ErrDuplicateURL
		fmt.Println("Existing short URL:", shortURL)
	}

	return shortURL, retErr
}

// GenerateShortID generates a unique identifier for a shortened URL.
func GenerateShortID() string {
	id, _ := nanoid.New()
	return id
}
