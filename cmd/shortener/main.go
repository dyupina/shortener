package main

import (
	// "io"
	"net/http"
	"shortener/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.ShortenURL)
	mux.HandleFunc("/{id}", handlers.GetOriginalURL)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
