package main

import (
	"net/http"

	"shortener/internal/app"
	"shortener/internal/config"
	"shortener/internal/handlers"
	"shortener/internal/logger"
	"shortener/internal/user"

	_ "net/http/pprof" //nolint:gosec // Use for Iter16

	"github.com/go-chi/chi/v5"
)

func main() {
	c := config.NewConfig()
	config.Init(c)

	s := app.SelectStorage(c)

	sugarLogger, err := logger.NewLogger()
	if err != nil {
		sugarLogger.Fatalf("Failed to initialize logger: %v", err)
	}

	userService := user.NewUserService()
	ctrl := handlers.NewController(c, s, sugarLogger, userService)
	r := chi.NewRouter()

	app.InitMiddleware(r, c, ctrl)
	app.Routing(r, ctrl)

	err = http.ListenAndServe(c.Addr, r) //nolint:gosec // Use chi Timeout (see above)
	if err != nil {
		sugarLogger.Fatalf("Failed to start server: %v", err)
	}
}
