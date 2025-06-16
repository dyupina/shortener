// Package models provides structure for storing URL information in JSON format.
package models

// StorageJSON - structure for storing URL information in JSON format.
type StorageJSON struct {
	// UUID: unique identifier for the URL record.
	UUID string `json:"uuid"`
	// ShortURL: shortened version of the URL associated with this record.
	ShortURL string `json:"short_url"`
	// OriginalURL: original URL that corresponds to the shortened version.
	OriginalURL string `json:"original_url"`
}

// StatsResponse represents the response for the /api/internal/stats endpoint.
type StatsResponse struct {
	// URLs: number of shortened URLs
	URLs int `json:"urls"`
	// Users: number of users
	Users int `json:"users"`
}

type BatchRequestEntity struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseEntity struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
