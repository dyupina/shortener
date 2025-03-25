package storage

import (
	"fmt"
	"net/http"
	"shortener/internal/repository"
	"sync"
)

// StorageMemory - структура для хранения данных URL в оперативной памяти.
type StorageMemory struct {
	urlStorage map[string]string
	mu         sync.Mutex
}

// NewStorageMemory создаёт и возвращает новый экземпляр StorageMemory.
func NewStorageMemory() *StorageMemory {
	return &StorageMemory{
		urlStorage: make(map[string]string),
	}
}

// UpdateData обновляет данные в хранилище и возвращает сокращённый URL.
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

// GetData извлекает оригинальный URL.
func (s *StorageMemory) GetData(shortID string) (originalURL string, isDeleted bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", false, fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, false, nil
}

// Ping проверяет соединение с базой данных. В данном случае не используется.
func (s *StorageMemory) Ping() error {
	return nil
}

// BatchDeleteURLs отмечает URL-адреса как удаленные в базе данных для заданного пользователя.
// В данном случае не используется.
func (s *StorageMemory) BatchDeleteURLs(userID string, urlIDs []string) error {
	return nil
}
