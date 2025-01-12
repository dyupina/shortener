package storage

import (
	"fmt"
	"shortener/internal/service"
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

func (s *StorageMemory) UpdateData(originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortID := service.GenerateShortID()

	for k, v := range s.urlStorage {
		if v == originalURL {
			return k, service.ErrDuplicateURL
		}
	}

	s.urlStorage[shortID] = originalURL
	newMap := make(map[string]string)
	newMap[shortID] = originalURL

	return shortID, nil
}

func (s *StorageMemory) GetData(shortID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, nil
}

func (s *StorageMemory) Ping() error {
	return nil
}
