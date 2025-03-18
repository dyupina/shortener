package storage

import (
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type StorageService interface {
	UpdateData(req *http.Request, originalURL, userID string) (shortURL string, retErr error)
	GetData(shortID string) (originalURL string, isDeleted bool, err error)
	Ping() error

	BatchDeleteURLs(userID string, urlIDs []string) error
}
