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

echo $SERVER_ADDRESS && echo $BASE_URL

unset SERVER_ADDRESS && unset BASE_URL
*/

package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

var (
	Addr    string
	BaseURL string
)

func Init() {
	// значения по-умолчанию
	Addr = "localhost:8080"
	BaseURL = "http://localhost:8080"

	if val, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		Addr = val
	}
	if val, exist := os.LookupEnv("BASE_URL"); exist {
		BaseURL = val
	}

	flag.StringVar(&Addr, "a", Addr, "адрес запуска HTTP-сервера")
	flag.StringVar(&BaseURL, "b", BaseURL, "базовый адрес результирующего сокращённого URL")

	// запускаем парсинг
	flag.Parse()

}
