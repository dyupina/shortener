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

	r.Get("/", handlers.GetOriginalURL)
	r.Post("/{id}", handlers.ShortenURL)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
