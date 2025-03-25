package storage

import (
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// StorageService описывает интерфейс для реализации разных типов хранилищ данных URL.
type StorageService interface {
	// UpdateData обновляет данные в хранилище и возвращает сокращённый URL.
	UpdateData(req *http.Request, originalURL, userID string) (shortURL string, retErr error)
	// GetData извлекает оригинальный URL.
	GetData(shortID string) (originalURL string, isDeleted bool, err error)
	// Ping проверяет соединение с базой данных, если она используется.
	Ping() error
	// BatchDeleteURLs отмечает URL-адреса как удаленные в базе данных для заданного пользователя,
	// если БД используется
	BatchDeleteURLs(userID string, urlIDs []string) error
}
