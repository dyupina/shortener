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
