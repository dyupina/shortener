package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	addr := config.Addr.Host + ":" + strconv.Itoa(config.Addr.Port)
	fmt.Printf(">>>>> %s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
