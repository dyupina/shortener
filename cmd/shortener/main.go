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

	dbConn, err := storage.InitializeDB(c)
	if err != nil {
		sugarLogger.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	ctrl := handlers.NewController(c, s, sugarLogger, dbConn)
	r := chi.NewRouter()

	s.RestoreURLstorage(c)

	file, err := storage.OpenFileAsWriter(c)
	if err != nil {
		sugarLogger.Fatalf("Failed to open URLs backup file: %v", err)
	}
	defer storage.ReadWriteCloserClose(file)
	s.AutoSave(file, c)

	controller.InitMiddleware(r, c, ctrl)
	controller.Routing(r, ctrl)

	err = http.ListenAndServe(c.Addr, r) //nolint:gosec // Use chi Timeout (see above)
	if err != nil {
		sugarLogger.Fatalf("Failed to start server: %v", err)
	}
}
