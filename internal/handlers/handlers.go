package handlers

import (
	"io"
	"net/http"
	"shortener/internal/config"

	"strings"

	"github.com/9ssi7/nanoid"
)

type stor interface {
	UpdateData(shortID, originalURL string)
	GetData(shortID string) (string, error)
}

type Controller struct {
	config *config.Config
	st     stor
}

func NewController(config *config.Config, st stor) *Controller {
	return &Controller{config: config, st: st}
}

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}

func (con *Controller) ShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		originalURL := string(body)
		shortID := generateShortID()

		con.st.UpdateData(shortID, originalURL)

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(con.config.BaseURL + "/" + shortID))
	}
}

func (con *Controller) GetOriginalURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/")
		originalURL, err := con.st.GetData(id)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}
