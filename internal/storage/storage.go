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

type Repository interface {
	GetData(key string) (string, error)
	UpdateData(key, value string) error
}

func (s *Storage) GetData(key string) (string, error) {
	value, exists := s.URLStorage[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

func (s *Storage) UpdateData(key, value string) error {
	s.URLStorage[key] = value
	//          short   orig
	return nil
}
