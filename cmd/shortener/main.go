package main

import (
	"log"
	"net/http"
	"time"

	"shortener/internal/config"
	"shortener/internal/handlers"

	"shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	c := config.NewConfig()
	config.Init(c)

	s := storage.NewURLstorage()

	controller := handlers.NewController(c, s)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Duration(c.Timeout) * time.Second))

	r.Post("/", handlers.WithLogging(c.Sugar, controller.ShortenURL()))
	r.Get("/{id}", handlers.WithLogging(c.Sugar, controller.GetOriginalURL()))

	err := http.ListenAndServe(c.Addr, r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
