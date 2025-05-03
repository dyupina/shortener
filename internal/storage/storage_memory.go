package storage

import (
	"fmt"
	"net/http"
	"shortener/internal/repository"
	"sync"
)

// StorageMemory - structure for storing URL data in memory.
type StorageMemory struct {
	urlStorage map[string]string
	mu         sync.Mutex
}

// NewStorageMemory creates and returns a new instance of StorageMemory.
func NewStorageMemory() *StorageMemory {
	return &StorageMemory{
		urlStorage: make(map[string]string),
	}
}

// UpdateData updates the data in the storage and returns the shortened URL.
func (s *StorageMemory) UpdateData(req *http.Request, originalURL, userID string) (shortURL string, retErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	retErr = nil

	shortURL = repository.GenerateShortID()

	for k, v := range s.urlStorage {
		if v == originalURL {
			return k, repository.ErrDuplicateURL
		}
	}

	s.urlStorage[shortURL] = originalURL
	newMap := make(map[string]string)
	newMap[shortURL] = originalURL

	return shortURL, nil
}

// GetData retrieves the original URL.
func (s *StorageMemory) GetData(shortID string) (originalURL string, isDeleted bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", false, fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, false, nil
}

// Ping checks the connection to the database. Not used in this context.
func (s *StorageMemory) Ping() error {
	return nil
}

// Close closes db connection. Not used in this case.
func (s *StorageMemory) Close() error {
	return nil
}

// BatchDeleteURLs marks URLs as deleted in the database for a specified user.
// Not used in this context.
func (s *StorageMemory) BatchDeleteURLs(userID string, urlIDs []string) error {
	return nil
}
