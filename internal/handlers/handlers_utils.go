package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"shortener/internal/config"
	"shortener/internal/storage"

	"github.com/9ssi7/nanoid"
	"go.uber.org/zap"
)

var shorturl struct {
	URL string `json:"result"`
}

var origurl struct {
	URL string `json:"url"`
}

type batchRequestEntity struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponseEntity struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

type Controller struct {
	conf  *config.Config
	st    storage.Storage
	sugar *zap.SugaredLogger
}

func (con *Controller) GetLogger() *zap.SugaredLogger {
	return con.sugar
}

func NewController(conf *config.Config, st storage.Storage, logger *zap.SugaredLogger) *Controller {
	return &Controller{
		conf:  conf,
		st:    st,
		sugar: logger,
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func extractURLfromHTML(res http.ResponseWriter, req *http.Request) string {
	b, _ := io.ReadAll(req.Body)
	body := string(b)

	re := regexp.MustCompile(`href=['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(body)

	if len(matches) > 1 {
		return matches[1]
	} else {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return ""
	}
}

func extractURLfromJSON(res http.ResponseWriter, req *http.Request) string {
	if err := json.NewDecoder(req.Body).Decode(&origurl); err != nil {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return ""
	}
	return origurl.URL
}

func extractURLsfromJSONBatchRequest(req *http.Request) []batchRequestEntity {
	var urls []batchRequestEntity
	err := json.NewDecoder(req.Body).Decode(&urls)
	if err != nil {
		return nil
	}
	return urls
}

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}
