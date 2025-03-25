// Package config используется для настройки конфигурации приложения.
package config

import (
	"flag"
	"os"
)

// Config - структура конфигурации приложения.
type Config struct {
	// Addr: строка с адресом, на котором будет запущен сервер (например, "localhost:8080").
	Addr string
	// BaseURL: базовый URL приложения, использующийся для формирования сокращённых ссылок.
	BaseURL string
	// Timeout: целочисленное значение времени (в секундах) для таймаута обработки запросов.
	Timeout int
	// URLStorageFile: путь к файлу, используемому для хранения URL-адресов.
	URLStorageFile string
	// DBConnection: строка подключения к базе данных.
	DBConnection string
	// NumWorkers: количество рабочих потоков, используемых приложением для обработки задач.
	NumWorkers int
}

// NewConfig создаёт и возвращает новый экземпляр структуры Config с предустановленными значениями.
func NewConfig() *Config {
	return &Config{
		Addr:           "localhost:8080",
		BaseURL:        "http://localhost:8080",
		Timeout:        15,
		URLStorageFile: "",
		DBConnection:   "",
		NumWorkers:     15,
	}
}

// Init инициализирует конфигурацию приложения, используя переменные окружения и флаги командной строки.
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

	flag.StringVar(&c.Addr, "a", c.Addr, "HTTP-server startup address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base address of the resulting shortened URL")
	flag.StringVar(&c.URLStorageFile, "f", c.URLStorageFile, "path to the file to save the data in JSON")
	flag.StringVar(&c.DBConnection, "d", c.DBConnection, "database connection address")

	flag.Parse()
}
