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
	r.Use(middleware.Timeout(time.Duration(c.Timeout) * time.Second))
	r.Use(controller.MiddlewareLogging)
	r.Use(controller.GzipEncodeMiddleware)
	r.Use(controller.GzipDecodeMiddleware)

	r.Post("/", controller.ShortenURL())
	r.Get("/{id}", controller.GetOriginalURL())
	r.Post("/api/shorten", controller.APIShortenURL())

	err := http.ListenAndServe(c.Addr, r) //nolint:gosec // Use chi Timeout (see above)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
