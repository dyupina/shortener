package handlers

import (
	"io"
	"net/http"

	"github.com/9ssi7/nanoid"
)

var urlStore = make(map[string]string)

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}

// curl -X POST http://localhost:8080 -H "Content-Type: text/plain" -d "https://example.com"
func ShortenURL(res http.ResponseWriter, req *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом POST
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(req.Body) // тело запроса
	originalURL := string(body)

	shortID := generateShortID()

	urlStore[shortID] = originalURL

	shortURL := "http://localhost:8080/" + shortID

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))

}

func GetOriginalURL(res http.ResponseWriter, req *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	id := req.PathValue("id") // получить id из /{id}

	originalURL, exists := urlStore[id]

	if !exists {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}
