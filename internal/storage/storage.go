package storage

import (
	"fmt"
	"shortener/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage interface {
	UpdateData(shortID, originalURL string)
	GetData(shortID string) (string, error)
	Ping() error
}

func SelectStorage(c *config.Config) Storage {
	if c.DBConnection != "" {
		fmt.Printf("try using DB\n")
		s := NewStorageDB(c)
		err := CreateTable(s)
		if err != nil {
			fmt.Printf(" database creation error\n")
		} else {
			return s
		}
	}

	if c.URLStorageFile != "" {
		fmt.Printf("try using file\n")
		s := NewStorageFile(c)
		err := RestoreURLstorage(c, s)
		if err != nil {
			fmt.Printf(" resore error\n")
		} else {
			AutoSave(s)
			return s
		}
	}

	fmt.Printf("using memory\n")
	s := NewStorageMemory()

	return s
}
