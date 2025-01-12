package storage

import (
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage interface {
	UpdateData(originalURL string) (string, error)
	GetData(shortID string) (string, error)
	Ping() error
}
