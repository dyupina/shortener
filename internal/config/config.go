package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var Addr *NetAddress

type NetAddress struct {
	Host    string
	Port    int
	BaseURL string
}

func Init() {
	addr := &NetAddress{ // значения по-умолчанию
		Host:    "localhost",
		Port:    8080,
		BaseURL: "http://localhost:8080",
	}

	// декларируем функцию-обработчик
	flag.Func("a", "адрес запуска HTTP-сервера host:port", func(flagValue string) error {
		hp := strings.Split(flagValue, ":")
		if len(hp) != 2 {
			return errors.New("need address in a form host:port")
		}
		port, err := strconv.Atoi(hp[1])
		if err != nil {
			return err
		}
		addr.Host = hp[0]
		addr.Port = port
		return nil
	})

	flag.Func("b", "базовый адрес результирующего сокращённого URL", func(flagValue string) error {
		addr.BaseURL = flagValue
		return nil
	})

	// запускаем парсинг
	flag.Parse()

	fmt.Println(addr.Host)
	fmt.Println(addr.Port)
	fmt.Println(addr.BaseURL)

	Addr = addr
}
