package config

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	// check default values
	require.Equal(t, "localhost:8080", config.Addr)
	require.Equal(t, "http://localhost:8080", config.BaseURL)
	require.Equal(t, 15, config.Timeout)
	require.Equal(t, "", config.URLStorageFile)
	require.Equal(t, "", config.DBConnection)
	require.Equal(t, 15, config.NumWorkers)
}

func TestInitWithEnvVariables(t *testing.T) {
	e1 := os.Setenv("SERVER_ADDRESS", "localhost:9090")
	e2 := os.Setenv("BASE_URL", "http://localhost:9090")
	e3 := os.Setenv("FILE_STORAGE_PATH", "/tmp/data")
	e4 := os.Setenv("DATABASE_DSN", "user:pass@/dbname")
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)

	defer func() {
		if e := os.Unsetenv("SERVER_ADDRESS"); e != nil {
			fmt.Println("os.Unsetenv(\"SERVER_ADDRESS\") error")
		}
	}()
	defer func() {
		if e := os.Unsetenv("BASE_URL"); e != nil {
			fmt.Println("os.Unsetenv(\"BASE_URL\") error")
		}
	}()
	defer func() {
		if e := os.Unsetenv("FILE_STORAGE_PATH"); e != nil {
			fmt.Println("os.Unsetenv(\"FILE_STORAGE_PATH\") error")
		}
	}()
	defer func() {
		if e := os.Unsetenv("DATABASE_DSN"); e != nil {
			fmt.Println("os.Unsetenv(\"DATABASE_DSN\") error")
		}
	}()

	config := NewConfig()
	err := Init(config)

	require.NoError(t, err)
	require.Equal(t, "localhost:9090", config.Addr)
	require.Equal(t, "http://localhost:9090", config.BaseURL)
	require.Equal(t, "/tmp/data", config.URLStorageFile)
	require.Equal(t, "user:pass@/dbname", config.DBConnection)
}

func TestInitWithFlags(t *testing.T) {
	args := []string{
		"-a", "127.0.0.1:8081",
		"-b", "http://127.0.0.1:8081",
		"-f", "/tmp/config.json",
		"-d", "postgres://user:pass@localhost/dbname",
	}

	oldArgs := os.Args
	os.Args = append([]string{oldArgs[0]}, args...)
	defer func() { os.Args = oldArgs }()

	config := NewConfig()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	err := Init(config)
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1:8081", config.Addr)
	require.Equal(t, "http://127.0.0.1:8081", config.BaseURL)
	require.Equal(t, "/tmp/config.json", config.URLStorageFile)
	require.Equal(t, "postgres://user:pass@localhost/dbname", config.DBConnection)
}
