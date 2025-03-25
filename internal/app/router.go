package app

import (
	"time"

	"shortener/internal/config"
	"shortener/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// InitMiddleware - инициализирует промежуточные обработчики (middleware) для маршрутизатора.
func InitMiddleware(r *chi.Mux, conf *config.Config, ctrl *handlers.Controller) {
	r.Use(ctrl.PanicRecoveryMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(conf.Timeout) * time.Second))
	r.Use(ctrl.Authenticate)
	r.Use(ctrl.LoggingMiddleware)
	r.Use(ctrl.GzipEncodeMiddleware)
	r.Use(ctrl.GzipDecodeMiddleware)
}

// Routing - регистрирует маршруты для работы с URL-контроллером.
// Зарегистрированные маршруты:
//   - POST "/": создаёт короткую версию URL с помощью ctrl.ShortenURL().
//   - GET "/{id}": возвращает оригинальный URL по сокращённой версии с использованием ctrl.GetOriginalURL().
//   - POST "/api/shorten": API-метод для сокращения URL через ctrl.APIShortenURL().
//   - POST "/api/shorten/batch": API-метод для пакетного сокращения URL через ctrl.APIShortenBatchURL().
//   - GET "/ping": проверка доступности сервиса через ctrl.PingHandler().
//   - GET "/api/user/urls": получение списка URL пользователя через ctrl.APIGetUserURLs().
//   - DELETE "/api/user/urls": удаление списка URL пользователя с помощью ctrl.DeleteUserURLs().
func Routing(r *chi.Mux, ctrl *handlers.Controller) {
	r.Post("/", ctrl.ShortenURL())
	r.Get("/{id}", ctrl.GetOriginalURL())
	r.Post("/api/shorten", ctrl.APIShortenURL())
	r.Post("/api/shorten/batch", ctrl.APIShortenBatchURL())
	r.Get("/ping", ctrl.PingHandler())
	r.Get("/api/user/urls", ctrl.APIGetUserURLs())
	r.Delete("/api/user/urls", ctrl.DeleteUserURLs())
}
