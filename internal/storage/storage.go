package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"shortener/internal/config"
)

type Storage struct {
	URLStorage map[string]string
}

type StorageJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *Storage) RestoreURLstorage(c *config.Config) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_RDONLY|os.O_CREATE, 0666) //nolint:mnd // read and write permission for all users
	if err != nil {
		fmt.Printf("error open file %s %s\n", c.URLStorageFile, err.Error())
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		urlFileStorage := StorageJSON{}
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &urlFileStorage); err != nil {
			fmt.Printf("error Unmarshal %s\n", err.Error())
			return
		}
		s.UpdateData(urlFileStorage.ShortURL, urlFileStorage.OriginalURL)
	}
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

func (s *Storage) GetStorageLen() int {
	return len(s.URLStorage)
}
