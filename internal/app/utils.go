package app

import (
	"log"
	"shortener/internal/config"
	"shortener/internal/storage"
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
