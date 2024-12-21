package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"regexp"
	"shortener/internal/config"
	"strconv"
	"time"

	"encoding/json"
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
	conf  *config.Config
	st    store
	sugar zap.SugaredLogger
}

func NewController(conf *config.Config, st store) *Controller {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}

	return &Controller{
		conf:  conf,
		st:    st,
		sugar: *logger.Sugar(), // регистратор SugaredLogger
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func generateShortID() string {
	id, _ := nanoid.New()
	return id
}

func (con *Controller) MiddlewareCompressing(next http.Handler) http.Handler {
	compressFn := func(res http.ResponseWriter, req *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не используется, передаём управление дальше без изменений
			next.ServeHTTP(res, req)
			return
		}

		// Функция сжатия должна работать для контента с типами application/json и text/html
		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") &&
			!strings.Contains(req.Header.Get("Content-Type"), "text/html") {
			// если типы не такие, сжатие не используем
			next.ServeHTTP(res, req)
			return
		}

		// Сжатие маленького тела (до 1400 байт)
		minSize := 1400
		contentLength, _ := strconv.Atoi(req.Header.Get("Content-Length"))
		if contentLength < minSize {
			// если размер меньше minSize байт, сжатие не используем
			next.ServeHTTP(res, req)
			return
		}

		// создаём gzip.Writer поверх текущего res
		gzip, err := gzip.NewWriterLevel(res, gzip.BestSpeed)
		if err != nil {
			http.Error(res, "Error creating gzip.Writer", http.StatusBadRequest)
			return
		}

		defer gzip.Close()

		res.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: res, Writer: gzip}, req)
	}

	return http.HandlerFunc(compressFn)
}

func (con *Controller) MiddlewareLogging(next http.Handler) http.Handler {
	logFn := func(res http.ResponseWriter, req *http.Request) {
		sugar := con.sugar
		start := time.Now()
		uri := req.RequestURI // эндпоинт
		method := req.Method  // метод запроса

		if method == http.MethodGet {
			next.ServeHTTP(res, req)      // обслуживание оригинального запроса
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
			next.ServeHTTP(&lw, req) // внедряем реализацию http.ResponseWriter
			sugar.Infoln(
				"status", responseData.status, // получаем перехваченный код статуса ответа
				"size", responseData.size, // получаем перехваченный размер ответа
			)
		}
	}

	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

func extractURLfromHTML(res http.ResponseWriter, req *http.Request) string {
	b, _ := io.ReadAll(req.Body)
	body := string(b)

	// Регулярное выражение для извлечения URL из href
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
	var origurl struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(req.Body).Decode(&origurl); err != nil {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return ""
	}
	return origurl.URL
}

func (con *Controller) ShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var originalURL string

		if strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			originalURL = extractURLfromJSON(res, req)
		}

		if strings.Contains(req.Header.Get("Content-Type"), "text/html") {
			originalURL = extractURLfromHTML(res, req)
		}
		if strings.Contains(req.Header.Get("Content-Type"), "text/plane") {
			b, _ := io.ReadAll(req.Body)
			originalURL = string(b)
		}
		// fmt.Printf("ShortenURL(): originalURL: %s\n", originalURL)

		shortID := generateShortID()

		con.st.UpdateData(shortID, originalURL)

		res.WriteHeader(http.StatusCreated)
		_, err := res.Write([]byte(con.conf.BaseURL + "/" + shortID))
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

func (con *Controller) APIShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		originalURL := extractURLfromJSON(res, req)
		shortID := generateShortID()

		con.st.UpdateData(shortID, originalURL)

		var shorturl struct {
			URL string `json:"result"`
		}
		shorturl.URL = con.conf.BaseURL + "/" + shortID

		resp, err := json.Marshal(shorturl)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write(resp)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

func (con *Controller) GetOriginalURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/")
		originalURL, err := con.st.GetData(id)

		// fmt.Printf("GetOriginalURL(): originalURL: %s\n", originalURL)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}
