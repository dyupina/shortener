package models

// StorageJSON - структура для хранения информации об URL в формате JSON.
type StorageJSON struct {
	// UUID: уникальный идентификатор для записи URL.
	UUID string `json:"uuid"`
	// ShortURL: сокращённая версия URL, связанная с данной записью.
	ShortURL string `json:"short_url"`
	// OriginalURL: оригинальный URL, который соответствует сокращённой версии.
	OriginalURL string `json:"original_url"`
}
