package main

import (
	"log"
	"net/http"

	"shortener/internal/config"
	"shortener/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	config.Init()

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Post("/", handlers.ShortenURL)
	r.Get("/{id}", handlers.GetOriginalURL)

	err := http.ListenAndServe(config.Cfg.Addr, r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
