package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"shortener/internal/config"
	"shortener/internal/domain/models"
	"strconv"
	"sync"
	"syscall"
)

type StorageFile struct {
	urlStorage map[string]string
	mu         sync.Mutex
	Events     chan map[string]string
	file       io.Writer
}

func NewStorageFile(c *config.Config) *StorageFile {
	bufSize := 100

	file, err := OpenFileAsWriter(c)
	if err != nil {
		fmt.Printf("Failed to open URLs backup file: %v", err)
	}

	return &StorageFile{
		urlStorage: make(map[string]string),
		Events:     make(chan map[string]string, bufSize),
		file:       file,
	}
}

func (s *StorageFile) UpdateData(shortID, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urlStorage[shortID] = originalURL
	newMap := make(map[string]string)
	newMap[shortID] = originalURL

	s.Events <- newMap
}

func (s *StorageFile) GetData(shortID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, nil
}

func RestoreURLstorage(c *config.Config, s *StorageFile) error {
	file, err := OpenFileAsReader(c)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Printf("Shutting down server and closing file...")
		ReadWriteCloserClose(file)
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		urlFileStorage := models.StorageJSON{}
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &urlFileStorage); err != nil {
			fmt.Printf("error Unmarshal %s\n", err.Error())
			return err
		}
		s.UpdateData(urlFileStorage.ShortURL, urlFileStorage.OriginalURL)
	}
	_ = os.Truncate(c.URLStorageFile, 0)
	return nil
}

func AutoSave(s *StorageFile) {
	go func() {
		i := 0
		for {
			newMap := <-s.Events
			BackupURLs(s, newMap, i+1)
			i++
		}
	}()
}

func BackupURLs(s *StorageFile, newMap map[string]string, counter int) {
	shortID := ""
	originalURL := ""

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range newMap {
		shortID = key
		originalURL = value

		urlFileStorage := models.StorageJSON{
			UUID:        strconv.Itoa(counter),
			ShortURL:    shortID,
			OriginalURL: originalURL,
		}

		data, err := json.Marshal(&urlFileStorage)
		if err != nil {
			fmt.Printf("error Marshal %s\n", err.Error())
			return
		}
		data = append(data, '\n')

		_, err = s.file.Write(data)
		if err != nil {
			fmt.Printf("error backup\n")
		}
	}
}

func (s *StorageFile) Ping() error {
	return nil
}

func OpenFileAsReader(c *config.Config) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_RDONLY|os.O_CREATE, 0666) //nolint:mnd // read and write permission for all users
	if err != nil {
		return nil, fmt.Errorf("error open file %s %s", c.URLStorageFile, err.Error())
	}
	return file, nil
}

func OpenFileAsWriter(c *config.Config) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_WRONLY|os.O_APPEND, 0666) //nolint:mnd // read and write permission for all users
	if err != nil {
		return nil, fmt.Errorf("error open file %s %s", c.URLStorageFile, err.Error())
	}
	return file, nil
}

func ReadWriteCloserClose(rwc io.ReadWriteCloser) {
	_ = rwc.Close()
}
