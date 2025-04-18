package storage

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"shortener/internal/config"
	"shortener/internal/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestNewStorageDB(t *testing.T) {
	t.Run("Successful connection and migration", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			if e := db.Close(); e != nil {
				fmt.Println("db.Close() error")
			}
		}()
		UpDBMigrations(db)
		storage := NewStorageDB("valid_connection_string")
		require.NotNil(t, storage)
		require.NoError(t, err, "Expected no error with valid connection string")
		require.NoError(t, mock.ExpectationsWereMet(), "Unfulfilled expectations")
	})

	t.Run("Empty connection string", func(t *testing.T) {
		storage := NewStorageDB("")
		require.NotNil(t, storage)
		require.NotNil(t, storage.DBConn, "Expected DBConn to be initialized even with empty connection string")
	})
}

func TestNewStorageFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_file")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if e := os.Remove(tmpFile.Name()); e != nil {
			fmt.Println("os.Remove(tmpFile.Name()) error")
		}
	}()

	// Mock
	config := &config.Config{
		URLStorageFile: tmpFile.Name(),
	}

	storageFile := NewStorageFile(config)
	require.NotNil(t, storageFile, "Expected non-nil StorageFile")
	require.NotNil(t, storageFile.file, "Expected file to be opened")
	require.NotNil(t, storageFile.urlStorage, "Expected urlStorage to be initialized")
	require.Equal(t, 100, cap(storageFile.Events), "Expected Events channel buffer size to be 100")
}

func TestNewStorageMemory(t *testing.T) {
	storageMemory := NewStorageMemory()
	require.NotNil(t, storageMemory, "Expected non-nil StorageMemory")
	require.NotNil(t, storageMemory.urlStorage, "Expected urlStorage map to be initialized")
}

func TestStorageMemory_Ping(t *testing.T) {
	storageMem := NewStorageMemory()
	err := storageMem.Ping()
	require.NoError(t, err)
}

func TestStorageFile_Ping(t *testing.T) {
	storageF := NewStorageFile(config.NewConfig())
	err := storageF.Ping()
	require.NoError(t, err)
}

func TestStorageDB_Ping(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.ExpectPing()
	require.NoError(t, err)
	defer func() {
		if e := db.Close(); e != nil {
			fmt.Println("db.Close() error")
		}
	}()

	storageDB := &StorageDB{DBConn: db}
	err = storageDB.Ping()
	require.NoError(t, err)
}

func TestStorageMemory_GetData(t *testing.T) {
	storage := &StorageMemory{
		urlStorage: map[string]string{
			"abc123": "http://example.com",
		},
	}

	t.Run("Existing shortID", func(t *testing.T) {
		originalURL, isDeleted, err := storage.GetData("abc123")
		require.NoError(t, err, "Expected no error for existing shortID")
		require.Equal(t, "http://example.com", originalURL, "Expected original URL to match")
		require.False(t, isDeleted, "Expected isDeleted to be false for existing shortID")
	})

	t.Run("Non-existing shortID", func(t *testing.T) {
		originalURL, isDeleted, err := storage.GetData("nonexistent")
		require.Error(t, err, "Expected error for non-existing shortID")
		require.EqualError(t, err, "shortID not found: nonexistent")
		require.Empty(t, originalURL, "Expected original URL to be empty for non-existing shortID")
		require.False(t, isDeleted, "Expected isDeleted to be false for non-existing shortID")
	})
}

func TestStorageFile_GetData(t *testing.T) {
	storage := &StorageFile{
		urlStorage: map[string]string{
			"abc123": "http://example.com",
		},
	}

	t.Run("Existing shortID", func(t *testing.T) {
		originalURL, isDeleted, err := storage.GetData("abc123")
		require.NoError(t, err, "Expected no error for existing shortID")
		require.Equal(t, "http://example.com", originalURL, "Expected original URL to match")
		require.False(t, isDeleted, "Expected isDeleted to be false for existing shortID")
	})

	t.Run("Non-existing shortID", func(t *testing.T) {
		originalURL, isDeleted, err := storage.GetData("nonexistent")
		require.Error(t, err, "Expected error for non-existing shortID")
		require.EqualError(t, err, "shortID not found: nonexistent")
		require.Empty(t, originalURL, "Expected original URL to be empty for non-existing shortID")
		require.False(t, isDeleted, "Expected isDeleted to be false for non-existing shortID")
	})
}

func TestStorageDB_GetData(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if e := db.Close(); e != nil {
			fmt.Println("db.Close() error")
		}
	}()

	storageDB := &StorageDB{DBConn: db}

	shortID := "shortURL123"
	expectedOriginalURL := "http://example.com"
	expectedIsDeleted := false

	t.Run("Existing shortID", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"original_url", "is_deleted"}).
			AddRow(expectedOriginalURL, expectedIsDeleted)
		mock.ExpectQuery("SELECT original_url, is_deleted FROM urls").
			WithArgs(shortID).WillReturnRows(rows)

		originalURL, isDeleted, err := storageDB.GetData(shortID)
		require.NoError(t, err, "Expected no error for existing shortID")
		require.Equal(t, expectedOriginalURL, originalURL, "Expected original URL to match")
		require.False(t, isDeleted, "Expected isDeleted to be false for existing shortID")
	})

	t.Run("Non-existing shortID", func(t *testing.T) {
		mock.ExpectQuery("SELECT original_url, is_deleted FROM urls").WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

		originalURL, isDeleted, err := storageDB.GetData("nonexistent")
		require.NoError(t, err, "Expected no error for non-existing shortID")
		require.Empty(t, originalURL, "Expected original URL to be empty for non-existing shortID")
		require.True(t, isDeleted, "Expected isDeleted to be true for non-existing shortID")
	})

	require.NoError(t, mock.ExpectationsWereMet(), "Unfulfilled expectations")
}

func TestStorageMemory_UpdateData(t *testing.T) {
	storage := &StorageMemory{
		urlStorage: make(map[string]string),
	}

	// Mock
	req, err := http.NewRequest("POST", "/", nil)
	require.NoError(t, err)

	t.Run("Add new original URL", func(t *testing.T) {
		shortURL, err := storage.UpdateData(req, "http://example.com", "user123")
		require.NoError(t, err, "Expected no error when adding new original URL")
		require.NotEmpty(t, shortURL, "Expected a non-empty short URL")
		require.Equal(t, "http://example.com", storage.urlStorage[shortURL], "Expected stored URL to match original")
	})

	t.Run("Add duplicate original URL", func(t *testing.T) {
		_, err := storage.UpdateData(req, "http://example.com", "user123")
		require.Error(t, err, "Expected error for duplicate original URL")
		require.Equal(t, repository.ErrDuplicateURL, err, "Expected duplicate URL error")
	})
}

func TestStorageFile_UpdateData(t *testing.T) {
	storage := &StorageFile{
		urlStorage: make(map[string]string),
		Events:     make(chan map[string]string, 1),
	}

	// Mock
	req, err := http.NewRequest("POST", "/", nil)
	require.NoError(t, err)

	t.Run("Add new original URL", func(t *testing.T) {
		shortURL, err := storage.UpdateData(req, "http://example.com", "user123")
		require.NoError(t, err, "Expected no error when adding new original URL")
		require.NotEmpty(t, shortURL, "Expected a non-empty short URL")
		require.Equal(t, "http://example.com", storage.urlStorage[shortURL], "Expected stored URL to match original")

		event := <-storage.Events
		require.Equal(t, "http://example.com", event[shortURL], "Expected event to contain the correct URL")
	})

	t.Run("Add duplicate original URL", func(t *testing.T) {
		_, err := storage.UpdateData(req, "http://example.com", "user123")
		require.Error(t, err, "Expected error for duplicate original URL")
		require.Equal(t, repository.ErrDuplicateURL, err, "Expected duplicate URL error")
	})
}
