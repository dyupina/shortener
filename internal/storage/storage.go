package storage

import (
	_ "github.com/jackc/pgx/v5/stdlib"
)

// StorageService describes the interface for implementing different types of URL data storage.
type StorageService interface {
	// UpdateData updates the data in the storage and returns the shortened URL.
	UpdateData(originalURL, userID string) (shortURL string, retErr error)
	// GetData retrieves the original URL.
	GetData(shortID string) (originalURL string, isDeleted bool, err error)
	// Ping checks the connection to the database, if one is used.
	Ping() error
	// Close closes db connection.
	Close() error
	// BatchDeleteURLs marks URLs as deleted in the database for a given user,
	// if a database is used.
	BatchDeleteURLs(userID string, urlIDs []string) error
}
