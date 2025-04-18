package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"shortener/internal/config"
	"shortener/internal/domain/models"
	"shortener/internal/repository"
	"strconv"
	"sync"
	"syscall"
)

// StorageFile - structure for storing URL data in a file.
type StorageFile struct {
	urlStorage map[string]string
	Events     chan map[string]string
	file       io.Writer
	mu         sync.Mutex
}

// NewStorageFile creates and returns a new instance of StorageFile.
func NewStorageFile(c *config.Config) *StorageFile {
	bufSize := 100

	file, err := OpenFileAsWriter(c)
	if err != nil {
		fmt.Printf("Failed to open URLs backup file: %v", err)
		return nil
	}

	return &StorageFile{
		urlStorage: make(map[string]string),
		Events:     make(chan map[string]string, bufSize),
		file:       file,
	}
}

// UpdateData updates the data in the storage and returns the shortened URL.
func (s *StorageFile) UpdateData(req *http.Request, originalURL, userID string) (shortURL string, retErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortURL = repository.GenerateShortID()

	for k, v := range s.urlStorage {
		if v == originalURL {
			return k, repository.ErrDuplicateURL
		}
	}

	newMap := make(map[string]string)
	newMap[shortURL] = originalURL

	s.urlStorage[shortURL] = originalURL

	s.Events <- newMap

	return shortURL, nil
}

// GetData retrieves the original URL and deletion status from the storage.
func (s *StorageFile) GetData(shortID string) (originalURL string, isDeleted bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", false, fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, false, nil
}

// RestoreURLstorage restores URL data from a backup file.
func RestoreURLstorage(c *config.Config, s *StorageFile) error {
	file, err := OpenFileAsReader(c)
	if err != nil {
		return err
	}
	defer ReadWriteCloserClose(file)

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

		s.urlStorage[urlFileStorage.ShortURL] = urlFileStorage.OriginalURL

		newMap := make(map[string]string)
		newMap[urlFileStorage.ShortURL] = urlFileStorage.OriginalURL
		s.Events <- newMap
	}
	_ = os.Truncate(c.URLStorageFile, 0)

	return nil
}

// AutoSave initiates automatic saving of URL data changes.
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

// BackupURLs performs backup of URL data to a file.
func BackupURLs(s *StorageFile, newMap map[string]string, counter int) {
	shortID := ""
	originalURL := ""
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

		s.mu.Lock()
		_, err = s.file.Write(data)
		s.mu.Unlock()

		if err != nil {
			fmt.Printf("error backup\n")
		}
	}
}

// Ping checks the connection to the database. Not used in this case.
func (s *StorageFile) Ping() error {
	return nil
}

// BatchDeleteURLs marks URLs as deleted in the database for a given user.
// Not used in this instance.
func (s *StorageFile) BatchDeleteURLs(userID string, urlIDs []string) error {
	return nil
}

// OpenFileAsReader opens a file for reading and creates the file if it does not exist.
func OpenFileAsReader(c *config.Config) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_RDONLY|os.O_CREATE, 0666) //nolint:mnd // read and write permission for all users
	if err != nil {
		return nil, fmt.Errorf("error open file %s %s", c.URLStorageFile, err.Error())
	}
	return file, nil
}

// OpenFileAsWriter opens a file for writing and creates the file if it does not exist.
func OpenFileAsWriter(c *config.Config) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666) //nolint:mnd // same
	if err != nil {
		return nil, fmt.Errorf("error open file %s %s", c.URLStorageFile, err.Error())
	}
	return file, nil
}

// ReadWriteCloserClose closes the ReadWriteCloser.
func ReadWriteCloserClose(rwc io.ReadWriteCloser) {
	_ = rwc.Close()
}
