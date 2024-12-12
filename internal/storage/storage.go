package storage

import (
	"fmt"
)

type Storage struct {
	URLStorage map[string]string
}

func NewURLstorage() *Storage {
	return &Storage{URLStorage: make(map[string]string)}
}

func (s *Storage) GetData(shortID string) (string, error) {
	originalURL, exists := s.URLStorage[shortID]
	if !exists {
		return "", fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, nil
}

func (s *Storage) UpdateData(shortID, originalURL string) {
	s.URLStorage[shortID] = originalURL
}
