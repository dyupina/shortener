package app

import (
	"log"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/storage"
	"time"

	"go.uber.org/zap"
)

// SelectStorage - selects the storage for saving URLs: database, file, or memory.
func SelectStorage(c *config.Config) storage.StorageService {
	if c.DBConnection != "" {
		log.Printf("try using DB\n")
		s := storage.NewStorageDB(c.DBConnection)
		return s
	}

	if c.URLStorageFile != "" {
		log.Printf("try using file\n")
		s := storage.NewStorageFile(c)
		if s != nil {
			err := storage.RestoreURLstorage(c, s)
			if err != nil {
				log.Printf(" restore error\n")
			} else {
				storage.AutoSave(s)
				return s
			}
		} else {
			log.Printf(" error using file")
		}
	}

	log.Printf("using memory\n")
	s := storage.NewStorageMemory()

	return s
}

// CreateServer creates and configures an HTTP server.
func CreateServer(c *config.Config, handler http.Handler, logger *zap.SugaredLogger) *http.Server {
	addr := c.Addr
	if c.EnableHTTPS {
		addr = "localhost:8443"
		c.Addr = addr
		c.BaseURL = "https://" + addr
		logger.Infof("Shortener at %s\n", c.Addr)
	} else {
		logger.Infof("Shortener at %s\n", c.Addr)
	}

	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 20 * time.Second,
	}
}
