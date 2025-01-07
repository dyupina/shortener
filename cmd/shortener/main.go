package main

import (
	"log"
	"net/http"

	"shortener/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.GetOriginalURL)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
