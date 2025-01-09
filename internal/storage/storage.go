package storage

import (
	"bufio"
	"context"
	"database/sql"
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
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	URLStorage map[string]string
	mu         sync.Mutex
	Events     chan map[string]string
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

func (s *Storage) RestoreURLstorage(c *config.Config) {
	file, err := OpenFileAsReader(c)
	if err != nil {
		return
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
			return
		}
		s.UpdateData(urlFileStorage.ShortURL, urlFileStorage.OriginalURL)
	}
	_ = os.Truncate(c.URLStorageFile, 0)
}

func (s *Storage) AutoSave(file io.Writer, c *config.Config) {
	go func() {
		i := 0
		for {
			newMap := <-s.Events
			s.BackupURLs(c, file, newMap, i+1)
			i++
		}
	}()
}

func (s *Storage) BackupURLs(c *config.Config, file io.Writer, newMap map[string]string, counter int) {
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

		_, _ = file.Write(data)
	}
}

func NewURLstorage() *Storage {
	bufSize := 100
	return &Storage{
		URLStorage: make(map[string]string),
		Events:     make(chan map[string]string, bufSize),
	}
}

func (s *Storage) GetData(shortID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, exists := s.URLStorage[shortID]
	if !exists {
		return "", fmt.Errorf("shortID not found: %s", shortID)
	}
	return originalURL, nil
}

func (s *Storage) UpdateData(shortID, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.URLStorage[shortID] = originalURL
	newMap := make(map[string]string)
	newMap[shortID] = originalURL

	s.Events <- newMap
}

func (s *Storage) Len() int {
	return len(s.URLStorage)
}

func InitDB(conf *config.Config) (*sql.DB, error) {
	// dbConn, err := pgx.Connect(context.Background(), conf.DBConnection)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to connect to the database: %v", err)
	// }

	// err = dbConn.Ping(context.Background())
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to ping the database: %v", err)
	// }

	dbConn, err := sql.Open("pgx", conf.DBConnection)
	if err != nil {
		return nil, fmt.Errorf("unable open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = dbConn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error verifying a connection to the database: %v", err)
	}

	return dbConn, nil
}
