package handlers

import (
	"io"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/storage"

	"github.com/9ssi7/nanoid"
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	ShortenURL(c config.Config, s storage.Storage) http.HandlerFunc
	GetOriginalURL(s storage.Storage) http.HandlerFunc
}

type Controller struct{}

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}

func (con *Controller) ShortenURL(c config.Config, s storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// этот обработчик принимает только запросы, отправленные методом POST
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(req.Body)
		originalURL := string(body)
		shortID := generateShortID()

		s.UpdateData(shortID, originalURL)

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(c.BaseURL + "/" + shortID))
	}

}

func (con *Controller) GetOriginalURL(s storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// этот обработчик принимает только запросы, отправленные методом GET
		if req.Method != http.MethodGet {
			http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		id := chi.URLParam(req, "id")
		// id := strings.TrimPrefix(req.URL.Path, "/")
		originalURL, err := s.GetData(id)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}
