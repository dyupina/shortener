package config

import (
	"flag"
	"os"
	"os/user"
)

type Config struct {
	Addr           string
	BaseURL        string
	Timeout        int
	URLStorageFile string
}

func NewConfig() *Config {
	usr, _ := user.Current()
	path := usr.HomeDir + "/shortener_storage"

	return &Config{
		Addr:           "localhost:8080",
		BaseURL:        "http://localhost:8080",
		Timeout:        15,
		URLStorageFile: path,
	}
}

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

	flag.StringVar(&c.Addr, "a", c.Addr, "HTTP-server startup address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base address of the resulting shortened URL")
	flag.StringVar(&c.URLStorageFile, "f", c.URLStorageFile, "path to the file to save the data in JSON")

	flag.Parse()
}
