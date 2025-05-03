package main

import (
	"shortener/internal/app"
	"shortener/internal/config"
	"shortener/internal/handlers"
	"shortener/internal/logger"
	"shortener/internal/user"

	"net/http"
	_ "net/http/pprof" //nolint:gosec // Use for Iter16

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func info(l *zap.SugaredLogger) {
	l.Infof("Build version: %s", buildVersion)
	l.Infof("Build date: %s", buildDate)
	l.Infof("Build commit: %s", buildCommit)
}

func main() {
	sugarLogger, err := logger.NewLogger()
	if err != nil {
		sugarLogger.Fatalf("Failed to initialize logger: %v", err)
	}

	c := config.NewConfig()
	err = config.Init(c)
	if err != nil {
		sugarLogger.Fatalf("Failed to initialize config: %v", err)
	}

	s := app.SelectStorage(c)

	info(sugarLogger)

	userService := user.NewUserService()
	ctrl := handlers.NewController(c, s, sugarLogger, userService)
	r := chi.NewRouter()

	app.InitMiddleware(r, c, ctrl)
	app.Routing(r, ctrl)

	server := app.CreateServer(c, r, sugarLogger)

	go func() {
		if c.EnableHTTPS {
			err = server.ListenAndServeTLS("https/localhost.crt", "https/localhost.key")
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			sugarLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	ctrl.HandleGracefulShutdown(server)
}
