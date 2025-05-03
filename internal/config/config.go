// Package config is used to configure the application settings.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
)

// Config - application configuration structure.
type Config struct {
	// Addr: string with the address on which the server will run (e.g., "localhost:8080").
	Addr string `json:"server_address"`
	// BaseURL: base URL of the application used to create shortened links.
	BaseURL string `json:"base_url"`
	// URLStorageFile: path to the file used for storing URLs.
	URLStorageFile string `json:"file_storage_path"`
	// DBConnection: database connection string.
	DBConnection string `json:"database_dsn"`
	// ConfigPath: path to configuration file.
	ConfigPath string
	// Timeout: integer value representing the request processing timeout in seconds.
	Timeout int
	// NumWorkers: number of worker threads used by the application for task processing.
	NumWorkers int
	// EnableHTTPS: is HTTPS connection enabled
	EnableHTTPS bool `json:"enable_https"`
}

var cfgDefault = Config{
	Addr:           "localhost:8080",
	BaseURL:        "http://localhost:8080",
	Timeout:        15,
	URLStorageFile: "",
	DBConnection:   "",
	NumWorkers:     15,
	EnableHTTPS:    false,
	ConfigPath:     "",
}

// NewConfig creates and returns a new instance of the Config structure with predefined values.
func NewConfig() *Config {
	return &cfgDefault
}

// ErrReadConfig - error reading json config.
var ErrReadConfig = errors.New("reading json config")

// ErrParseConfig - error parsing json config.
var ErrParseConfig = errors.New("parse json config")

// Init initializes the application configuration using environment variables and command-line flags.
func Init(c *Config) error {
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

	var flagCgf Config
	flag.StringVar(&flagCgf.Addr, "a", "", "HTTP-server startup address")
	flag.StringVar(&flagCgf.BaseURL, "b", "", "base address of the resulting shortened URL")
	flag.StringVar(&flagCgf.URLStorageFile, "f", "", "path to the file to save the data in JSON")
	flag.StringVar(&flagCgf.DBConnection, "d", "", "database connection address")
	flag.BoolVar(&flagCgf.EnableHTTPS, "s", false, "is HTTPS connection enabled")
	flag.StringVar(&flagCgf.ConfigPath, "c", "", "path to config file (json)")

	flag.Parse()

	if flagCgf.ConfigPath != "" {
		file, err := os.ReadFile(flagCgf.ConfigPath)
		if err != nil {
			return ErrReadConfig
		}
		if err := json.Unmarshal(file, &c); err != nil {
			return ErrParseConfig
		}
	}

	// override
	if flagCgf.Addr != "" {
		c.Addr = flagCgf.Addr
	}
	if flagCgf.BaseURL != "" {
		c.BaseURL = flagCgf.BaseURL
	}
	if flagCgf.URLStorageFile != "" {
		c.URLStorageFile = flagCgf.URLStorageFile
	}
	if flagCgf.DBConnection != "" {
		c.DBConnection = flagCgf.DBConnection
	}
	if flagCgf.EnableHTTPS {
		c.EnableHTTPS = flagCgf.EnableHTTPS
	}

	return nil
}
