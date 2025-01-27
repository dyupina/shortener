package storage

import (
	"database/sql"
	"embed"
	"log"
	"shortener/internal/service"

	"github.com/pressly/goose/v3"
)

type StorageDB struct {
	DBConn *sql.DB
}

//go:embed db/migrations/*.sql
var embedMigrations embed.FS

func NewStorageDB(connetion string) *StorageDB {
	DBConn, _ := sql.Open("pgx", connetion)

	if connetion != "" {
		goose.SetBaseFS(embedMigrations)

		if err := goose.SetDialect("postgres"); err != nil {
			log.Printf("error setting SQL dialect\n")
		}

		if err := goose.Up(DBConn, "db/migrations"); err != nil {
			log.Printf("error migration %s\n", err.Error())
		}
	}

	return &StorageDB{
		DBConn: DBConn,
	}
}

func (s *StorageDB) UpdateData(originalURL string) (string, error) {
	var shortURL string
	var retErr error
	var serv = &service.Serv{}
	shortURL, retErr = serv.GetShortURLDB(originalURL, s.DBConn)

	return shortURL, retErr
}

const selectRow = "SELECT full_url FROM urls WHERE short_url=$1"

func (s *StorageDB) GetData(shortID string) (string, error) {
	var originalURL string
	err := s.DBConn.QueryRow(selectRow, shortID).Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (s *StorageDB) Ping() error {
	return s.DBConn.Ping()
}
