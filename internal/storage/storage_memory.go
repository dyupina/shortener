package storage

import (
	"fmt"
	"net/http"
	"shortener/internal/repository"
	"sync"
)

type StorageMemory struct {
	urlStorage map[string]string
	mu         sync.Mutex
}

func NewStorageMemory() *StorageMemory {
	return &StorageMemory{
		urlStorage: make(map[string]string),
	}
}

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

func (s *StorageMemory) GetData(shortID string) (originalURL string, isDeleted bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", false, fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, false, nil
}

func (s *StorageMemory) Ping() error {
	return nil
}

func (s *StorageMemory) BatchDeleteURLs(userID string, urlIDs []string) error {
	return nil
}
