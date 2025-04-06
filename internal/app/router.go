package app

import (
	"time"

	"shortener/internal/config"
	"shortener/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// InitMiddleware - initializes middleware handlers for the router.
func InitMiddleware(r *chi.Mux, conf *config.Config, ctrl *handlers.Controller) {
	r.Use(ctrl.PanicRecoveryMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(conf.Timeout) * time.Second))
	r.Use(ctrl.Authenticate)
	r.Use(ctrl.LoggingMiddleware)
	r.Use(ctrl.GzipEncodeMiddleware)
	r.Use(ctrl.GzipDecodeMiddleware)
	r.Mount("/debug", middleware.Profiler())
}

// Routing - registers routes for the URL controller.
// Registered routes:
//   - POST "/": creates a shortened version of a URL using ctrl.ShortenURL().
//   - GET "/{id}": returns the original URL from the shortened version using ctrl.GetOriginalURL().
//   - POST "/api/shorten": API method for shortening a URL through ctrl.APIShortenURL().
//   - POST "/api/shorten/batch": API method for batch URL shortening through ctrl.APIShortenBatchURL().
//   - GET "/ping": service availability check through ctrl.PingHandler().
//   - GET "/api/user/urls": retrieves the user's URL list through ctrl.APIGetUserURLs().
//   - DELETE "/api/user/urls": deletes the user's URL list using ctrl.DeleteUserURLs().
func Routing(r *chi.Mux, ctrl *handlers.Controller) {
	r.Post("/", ctrl.ShortenURL())
	r.Get("/{id}", ctrl.GetOriginalURL())
	r.Post("/api/shorten", ctrl.APIShortenURL())
	r.Post("/api/shorten/batch", ctrl.APIShortenBatchURL())
	r.Get("/ping", ctrl.PingHandler())
	r.Get("/api/user/urls", ctrl.APIGetUserURLs())
	r.Delete("/api/user/urls", ctrl.DeleteUserURLs())
}
