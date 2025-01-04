package main

import (
	"net/http"

	controller "shortener/internal/app"
	"shortener/internal/config"
	"shortener/internal/handlers"
	"shortener/internal/logger"
	"shortener/internal/storage"

	"github.com/go-chi/chi/v5"
)

func main() {
	c := config.NewConfig()
	config.Init(c)

	s := storage.NewURLstorage()

	sugarLogger, err := logger.NewLogger()
	if err != nil {
		sugarLogger.Fatalf("Failed to initialize logger: %v", err)
	}
	ctrl := handlers.NewController(c, s, sugarLogger)
	r := chi.NewRouter()

	s.RestoreURLstorage(c)

	file, err := storage.OpenFileAsWriter(c)
	if err != nil {
		return
	}
	defer storage.CloseWriter(file)
	s.AutoSave(file, c)

	controller.InitMiddleware(r, c, ctrl)
	controller.Routing(r, ctrl)

	err = http.ListenAndServe(c.Addr, r) //nolint:gosec // Use chi Timeout (see above)
	if err != nil {
		sugarLogger.Fatalf("Failed to start server: %v", err)
	}
}
