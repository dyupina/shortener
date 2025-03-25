package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/9ssi7/nanoid"
)

// ErrDuplicateURL - ошибка, когда оригинальный URL уже существует в системе.
var ErrDuplicateURL = errors.New("duplicate URL")

// Repository - интерфейс для работы с сокращёнными URL.
type Repository interface {
	GetShortURL_db(originalURL string) (string, error)
}

// Repo - структура для работы с хранилищем данных.
type Repo struct {
}

const insertRow = `
INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3)
ON CONFLICT (original_url) DO NOTHING
RETURNING short_url`

// GetShortURLDB возвращает сокращённый URL для заданного исходного URL и идентификатора пользователя.
// Если URL уже существует, возвращает существующий сокращённый URL с ошибкой ErrDuplicateURL.
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

// GenerateShortID генерирует уникальный идентификатор для сокращённого URL.
func GenerateShortID() string {
	id, _ := nanoid.New()
	return id
}
