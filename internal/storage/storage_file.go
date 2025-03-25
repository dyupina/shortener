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

// StorageFile - структура для хранения данных URL в файле.
type StorageFile struct {
	urlStorage map[string]string
	mu         sync.Mutex
	Events     chan map[string]string
	file       io.Writer
}

// NewStorageFile создаёт и возвращает новый экземпляр StorageFile.
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

// UpdateData обновляет данные в хранилище и возвращает сокращённый URL.
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

// GetData извлекает оригинальный URL и статус удаления из хранилища.
func (s *StorageFile) GetData(shortID string) (originalURL string, isDeleted bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.urlStorage[shortID]
	if !exists {
		return "", false, fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, false, nil
}

// RestoreURLstorage восстанавливает данные URL из резервной копии файла.
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

// AutoSave запускает автоматическое сохранение изменений данных URL.
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

// BackupURLs производит резервное копирование данных URL в файл.
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

// Ping проверяет соединение с базой данных. В данном случае не используется.
func (s *StorageFile) Ping() error {
	return nil
}

// BatchDeleteURLs отмечает URL-адреса как удаленные в базе данных для заданного пользователя.
// В данном случае не используется.
func (s *StorageFile) BatchDeleteURLs(userID string, urlIDs []string) error {
	return nil
}

// OpenFileAsReader открывает файл для чтения и создаёт файл, если его не существует.
func OpenFileAsReader(c *config.Config) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_RDONLY|os.O_CREATE, 0666) //nolint:mnd // read and write permission for all users
	if err != nil {
		return nil, fmt.Errorf("error open file %s %s", c.URLStorageFile, err.Error())
	}
	return file, nil
}

// OpenFileAsWriter открывает файл для записи и создаёт файл, если его не существует.
func OpenFileAsWriter(c *config.Config) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(c.URLStorageFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666) //nolint:mnd // same
	if err != nil {
		return nil, fmt.Errorf("error open file %s %s", c.URLStorageFile, err.Error())
	}
	return file, nil
}

// ReadWriteCloserClose закрывает ReadWriteCloser.
func ReadWriteCloserClose(rwc io.ReadWriteCloser) {
	_ = rwc.Close()
}
