package config

import (
	"flag"
)

var (
	Addr    string
	BaseURL string
)

func Init() {

	flag.StringVar(&Addr, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&BaseURL, "b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL")

	// запускаем парсинг
	flag.Parse()

}
