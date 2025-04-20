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

	if c.EnableHTTPS {
		httpsAddr := "localhost:8443"
		c.Addr = httpsAddr
		c.BaseURL = "https://" + httpsAddr
		sugarLogger.Infof("Shortener at %s\n", c.Addr)
		err = http.ListenAndServeTLS(httpsAddr, "https/localhost.crt", "https/localhost.key", r) //nolint:gosec // Use chi Timeout (see above)
	} else {
		sugarLogger.Infof("Shortener at %s\n", c.Addr)
		err = http.ListenAndServe(c.Addr, r) //nolint:gosec // Use chi Timeout (see above)
	}

	if err != nil {
		sugarLogger.Fatalf("Failed to start server: %v", err)
	}
}
