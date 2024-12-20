package config

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

type Config struct {
	Addr    string
	BaseURL string
	Timeout int
	Sugar   zap.SugaredLogger
}

func NewConfig() *Config {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}

	return &Config{
		Addr:    "localhost:8080",
		BaseURL: "http://localhost:8080",
		Timeout: 15,
		Sugar:   *logger.Sugar(), // регистратор SugaredLogger
	}
}

func Init(c *Config) {
	if val, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		c.Addr = val
	}
	if val, exist := os.LookupEnv("BASE_URL"); exist {
		c.BaseURL = val
	}

	flag.StringVar(&c.Addr, "a", c.Addr, "адрес запуска HTTP-сервера")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "базовый адрес результирующего сокращённого URL")

	flag.Parse()
}
