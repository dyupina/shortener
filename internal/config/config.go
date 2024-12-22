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

// func (c *Config) GetURLStorageFile() string {
// 	return c.URLStorageFile
// }

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

	flag.StringVar(&c.Addr, "a", c.Addr, "адрес запуска HTTP-сервера")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "базовый адрес результирующего сокращённого URL")
	flag.StringVar(&c.URLStorageFile, "f", c.URLStorageFile, "путь до файла, куда сохраняются данные в формате JSON")

	flag.Parse()
}
