// Package config is used to configure the application settings.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Config - application configuration structure.
type Config struct {
	// Addr: string with the address on which the server will run (e.g., "localhost:8080").
	Addr string
	// BaseURL: base URL of the application used to create shortened links.
	BaseURL string
	// URLStorageFile: path to the file used for storing URLs.
	URLStorageFile string
	// DBConnection: database connection string.
	DBConnection string
	// Timeout: integer value representing the request processing timeout in seconds.
	Timeout int
	// NumWorkers: number of worker threads used by the application for task processing.
	NumWorkers int
	// EnableHTTPS: is HTTPS connection enabled
	EnableHTTPS bool
}

// NewConfig creates and returns a new instance of the Config structure with predefined values.
func NewConfig() *Config {
	return &Config{
		Addr:           "localhost:8080",
		BaseURL:        "http://localhost:8080",
		Timeout:        15,
		URLStorageFile: "",
		DBConnection:   "",
		NumWorkers:     15,
		EnableHTTPS:    false,
	}
}

// Init initializes the application configuration using environment variables and command-line flags.
func Init(c *Config) {
	if val, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		c.Addr = val
	}
	if val, exist := os.LookupEnv("BASE_URL"); exist {
		c.BaseURL = val
	}
	if val, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		c.URLStorageFile = val
	}
	if val, exist := os.LookupEnv("DATABASE_DSN"); exist {
		c.DBConnection = val
	}
	if val, exist := os.LookupEnv("ENABLE_HTTPS"); exist {
		valBool, err := strconv.ParseBool(val)
		if err == nil {
			c.EnableHTTPS = valBool
		}
	}

	flag.StringVar(&c.Addr, "a", c.Addr, "HTTP-server startup address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base address of the resulting shortened URL")
	flag.StringVar(&c.URLStorageFile, "f", c.URLStorageFile, "path to the file to save the data in JSON")
	flag.StringVar(&c.DBConnection, "d", c.DBConnection, "database connection address")
	flag.BoolVar(&c.EnableHTTPS, "s", c.EnableHTTPS, "is HTTPS connection enabled")

	flag.Parse()
}
