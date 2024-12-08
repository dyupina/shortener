package storage

import (
	"fmt"
)

type Storage struct {
	URL_Storage map[string]string
}

func NewURLstorage() *Storage {
	return &Storage{URL_Storage: make(map[string]string)}
}

type Repository interface {
	GetData(key string) (string, error)
	UpdateData(key, value string) error
}

func (s *Storage) GetData(key string) (string, error) {
	value, exists := s.URL_Storage[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

func (s *Storage) UpdateData(key, value string) error {
	s.URL_Storage[key] = value
	return nil
}
