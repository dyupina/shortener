package app

import (
	"time"

	"shortener/internal/config"
	"shortener/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitMiddleware(r *chi.Mux, conf *config.Config, ctrl *handlers.Controller) {
	r.Use(ctrl.PanicRecoveryMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(conf.Timeout) * time.Second))
	r.Use(ctrl.Authenticate)
	r.Use(ctrl.LoggingMiddleware)
	r.Use(ctrl.GzipEncodeMiddleware)
	r.Use(ctrl.GzipDecodeMiddleware)
}

func Routing(r *chi.Mux, ctrl *handlers.Controller) {
	r.Post("/", ctrl.ShortenURL())
	r.Get("/{id}", ctrl.GetOriginalURL())
	r.Post("/api/shorten", ctrl.APIShortenURL())
	r.Post("/api/shorten/batch", ctrl.APIShortenBatchURL())
	r.Get("/ping", ctrl.PingHandler())
	r.Get("/api/user/urls", ctrl.APIGetUserURLs())
	r.Delete("/api/user/urls", ctrl.DeleteUserURLs())
}
