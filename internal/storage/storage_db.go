package storage

import (
	"database/sql"
	"embed"
	"log"
	"net/http"
	"shortener/internal/repository"

	"github.com/pressly/goose/v3"
)

// StorageDB - structure for working with a PostgreSQL database.
type StorageDB struct {
	DBConn *sql.DB
}

//go:embed db/migrations/*.sql
var embedMigrations embed.FS

// UpDBMigrations applies migrations to the database using migration files.
func UpDBMigrations(db *sql.DB) {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Printf("error setting SQL dialect\n")
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		log.Printf("error migration %s\n", err.Error())
	}
}

// NewStorageDB creates and returns a new instance of StorageDB storage with a connection to the database.
func NewStorageDB(connetion string) *StorageDB {
	DBConn, _ := sql.Open("pgx", connetion)

	if connetion != "" {
		UpDBMigrations(DBConn)
	}

	return &StorageDB{
		DBConn: DBConn,
	}
}

// UpdateData updates data in the storage and returns the shortened URL.
func (s *StorageDB) UpdateData(req *http.Request, originalURL, userID string) (shortURL string, retErr error) {
	var repo = &repository.Repo{}
	shortURL, retErr = repo.GetShortURLDB(userID, originalURL, s.DBConn)
	return shortURL, retErr
}

const updateSetIsDeleted = `UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2::text[])`
const selectFullURLAndIsDeleted = "SELECT original_url, is_deleted FROM urls WHERE short_url=$1"

// GetData retrieves the original URL and deletion status from the storage.
func (s *StorageDB) GetData(shortID string) (originalURL string, isDeleted bool, err error) {
	err = s.DBConn.QueryRow(selectFullURLAndIsDeleted, shortID).Scan(&originalURL, &isDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", true, nil // Если запись не найдена, можно считать ее удаленной
		}
		return "", false, err
	}
	return originalURL, isDeleted, nil
}

// Ping checks the connection to the database.
func (s *StorageDB) Ping() error {
	return s.DBConn.Ping()
}

// BatchDeleteURLs marks URLs as deleted in the database for a given user.
func (s *StorageDB) BatchDeleteURLs(userID string, urlIDs []string) error {
	_, err := s.DBConn.Exec(updateSetIsDeleted, userID, urlIDs)

	return err
}

// Close closes db connection.
func (s *StorageDB) Close() error {
	return s.DBConn.Close()
}
