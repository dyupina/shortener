package handlers

import (
	"io"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/storage"
	"strings"

	"github.com/9ssi7/nanoid"
)

// var urlStore = make(map[string]string)

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}

// curl -X POST http://localhost:8080 -H "Content-Type: text/plain" -d "https://example.com"
func ShortenURL(c config.Config, s storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// этот обработчик принимает только запросы, отправленные методом POST
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(req.Body) // тело запроса
		originalURL := string(body)

		shortID := generateShortID()

		s.UpdateData(shortID, originalURL)
		// s.URL_Storage[shortID] = originalURL

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(c.BaseURL + "/" + shortID))
	}

}

func GetOriginalURL(s storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// этот обработчик принимает только запросы, отправленные методом GET
		if req.Method != http.MethodGet {
			http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		// получить id из /{id}
		// id := req.PathValue("id")
		id := strings.TrimPrefix(req.URL.Path, "/")

		// originalURL, exists := s.URL_Storage[id]
		originalURL, err := s.GetData(id)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}
