package handlers

import (
	"io"
	"log"
	"net/http"
	"shortener/internal/config"
	"time"

	"strings"

	"github.com/9ssi7/nanoid"
	"go.uber.org/zap"
)

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

type store interface {
	UpdateData(shortID, originalURL string)
	GetData(shortID string) (string, error)
}

type Controller struct {
	conf *config.Config
	st   store
}

func NewController(conf *config.Config, st store) *Controller {
	return &Controller{conf: conf, st: st}
}

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}

func WithLogging(sugar zap.SugaredLogger, h http.Handler) http.HandlerFunc {
	logFn := func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		uri := req.RequestURI // эндпоинт
		method := req.Method  // метод запроса

		if method == http.MethodGet {
			h.ServeHTTP(res, req)         // обслуживание оригинального запроса
			duration := time.Since(start) // время выполнения запроса
			// отправляем сведения о запросе в zap
			sugar.Infoln(
				"uri", uri,
				"method", method,
				"duration", duration,
			)
		}

		if method == http.MethodPost {
			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: res, // встраиваем оригинальный http.ResponseWriter
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, req) // внедряем реализацию http.ResponseWriter
			sugar.Infoln(
				"status", responseData.status, // получаем перехваченный код статуса ответа
				"size", responseData.size, // получаем перехваченный размер ответа
			)
		}
	}

	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

func (con *Controller) ShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		originalURL := string(body)
		shortID := generateShortID()

		con.st.UpdateData(shortID, originalURL)

		res.WriteHeader(http.StatusCreated)
		_, err := res.Write([]byte(con.conf.BaseURL + "/" + shortID))
		if err != nil {
			log.Print("Error writing short URL") // ????????
		}
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
