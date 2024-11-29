/*
Добавьте возможность конфигурировать сервис с помощью переменных окружения:
Адрес запуска HTTP-сервера — с помощью переменной SERVER_ADDRESS.
Базовый адрес результирующего сокращённого URL — с помощью переменной BASE_URL.

Приоритет параметров сервера должен быть таким:
Если указана переменная окружения, то используется она.
Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.
*/

/*
export SERVER_ADDRESS=localhost:5555
export BASE_URL=http://localhost:5555
*/

package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

var Cfg Config

type Config struct {
	Addr    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func Init() {
	var cfg Config

	// значения по-умолчанию
	cfg.Addr = "lo55555calhost:8080"
	cfg.BaseURL = "http://localhost:8080"

	if val, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		cfg.Addr = val
	}
	if val, exist := os.LookupEnv("BASE_URL"); exist {
		cfg.BaseURL = val
	}
	Cfg = cfg

	err := env.Parse(&Cfg)

	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&Cfg.Addr, "a", Cfg.Addr, "адрес запуска HTTP-сервера")
	flag.StringVar(&Cfg.BaseURL, "b", Cfg.BaseURL, "базовый адрес результирующего сокращённого URL")

	// запускаем парсинг
	flag.Parse()

}
