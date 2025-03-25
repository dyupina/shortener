package handlers

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/repository"
	"shortener/internal/storage"
	"shortener/internal/user"
	"strconv"
	"time"

	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Controller управляет HTTP-запросами для работы с сокращением URL.
type Controller struct {
	conf           *config.Config
	storageService storage.StorageService
	sugar          *zap.SugaredLogger
	userService    user.UserService
}

// NewController создаёт и возвращает новый экземпляр Controller, используя переданные компоненты
// конфигурации, хранилища, логгера и сервиса пользователя.
func NewController(conf *config.Config, storageService storage.StorageService, logger *zap.SugaredLogger, us user.UserService) *Controller {
	return &Controller{
		conf:           conf,
		storageService: storageService,
		sugar:          logger,
		userService:    us,
	}
}

// DeleteUserURLs обрабатывает HTTP-запросы для удаления URL-адресов, принадлежащих пользователю.
// Поток удаленных URL-адресов обрабатывается асинхронно.
func (con *Controller) DeleteUserURLs() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userID := req.Header.Get("User-ID")
		if userID == "" {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var urlIDs []string
		err := json.NewDecoder(req.Body).Decode(&urlIDs)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		doneCh := make(chan struct{})

		inputCh := createURLBatchChannel(doneCh, urlIDs)
		workerChs := distributeDeleteTasks(doneCh, inputCh, con.conf.NumWorkers, userID, con)
		resultCh := collectDeletionResults(workerChs...)

		go func() {
			for res := range resultCh {
				con.sugar.Infof(" Deleted short URL: %s\n", res)
			}
		}()

		res.WriteHeader(http.StatusAccepted)
	}
}

// APIGetUserURLs обрабатывает запросы на получение всех URL-адресов, связанных с пользователем.
// Возвращает JSON-ответ с URL-адресами пользователя.
//
// HTTP-ответ:
//   - 401 Unauthorized: если пользователь не аутентифицирован.
//   - 204 No Content: если пользователь не имеет связанных URL-адресов.
//   - 200 OK: успешное получение URL-адресов пользователя в формате JSON.
//   - 500 Internal Server Error: если соединение не удалось.
func (con *Controller) APIGetUserURLs() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userID := req.Header.Get("User-ID")
		if userID == "" {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		res.Header().Set("Content-Type", "application/json")

		urls, exist := con.userService.GetUserURLs(userID)

		if !exist {
			con.sugar.Debug("(APIGetUserURLs) StatusUnauthorized userID %s\n", userID)
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		if len(urls) == 0 {
			con.sugar.Debug("(APIGetUserURLs) StatusNoContent")
			res.WriteHeader(http.StatusNoContent)
			return
		}

		response, err := json.Marshal(urls)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = res.Write(response)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

// Authenticate осуществляет аутентификацию пользователя с использованием HTTP-куки.
// Устанавливает новый идентификатор пользователя, если куки отсутствуют или недействительны.
func (con *Controller) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		uidFromCookie, err := con.userService.GetUserIDFromCookie(req)

		if err != nil || uidFromCookie == "" {
			con.sugar.Debugf("(Authenticate) Missing or invalid cookie: %s", err)

			uid := uuid.New().String()
			if err := con.userService.SetUserIDCookie(res, uid); err != nil {
				con.sugar.Errorf("(Authenticate) Failed to set user ID cookie: %s", err.Error())
				http.Error(res, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			con.userService.InitUserURLs(uid)
			con.sugar.Debugf("(Authenticate) New user ID set in cookie: %s", uid)
			req.Header.Set("User-ID", uid)
		} else {
			con.sugar.Debugf("(Authenticate) Valid user ID from cookie: %s", uidFromCookie)
			req.Header.Set("User-ID", uidFromCookie)
		}

		next.ServeHTTP(res, req)
	})
}

// GzipDecodeMiddleware декодирует входящее содержимое HTTP-запросов, закодированное с использованием gzip.
//
// HTTP-ответ:
//   - 400 Bad Request: если декодирование gzip-содержимого не удалось.
func (con *Controller) GzipDecodeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(res, "Bad Request: Unable to decode gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			req.Body = gz
		}
		next.ServeHTTP(res, req)
	})
}

// GzipEncodeMiddleware сжимает исходящие HTTP-ответы с использованием gzip.
// Условие сжатия: клиент поддерживает gzip (Accept-Encoding),
// а содержимое ответа является JSON или HTML, и его размер превышает минимальную границу для сжатия.
//
// HTTP-ответ:
//   - 400 Bad Request: если создание gzip.Writer не удалось.
func (con *Controller) GzipEncodeMiddleware(next http.Handler) http.Handler {
	compressFn := func(res http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(res, req)
			return
		}

		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") &&
			!strings.Contains(req.Header.Get("Content-Type"), "text/html") {
			next.ServeHTTP(res, req)
			return
		}

		minSize := 1400
		contentLength, _ := strconv.Atoi(req.Header.Get("Content-Length"))
		if contentLength < minSize {
			next.ServeHTTP(res, req)
			return
		}

		gzip, err := gzip.NewWriterLevel(res, gzip.BestSpeed)
		if err != nil {
			http.Error(res, "Error creating gzip.Writer", http.StatusBadRequest)
			return
		}

		defer gzip.Close()

		res.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(gzipWriter{ResponseWriter: res, Writer: gzip}, req)
	}
	return http.HandlerFunc(compressFn)
}

// LoggingMiddleware логирует информацию о HTTP-запросах и ответах.
//
// Логирует:
//   - Метод запроса (GET, POST, DELETE).
//   - URI запроса.
//   - Длительность обработки запроса.
//   - Статус и размер ответа в случае POST и DELETE.
func (con *Controller) LoggingMiddleware(next http.Handler) http.Handler {
	logFn := func(res http.ResponseWriter, req *http.Request) {
		sugar := con.sugar
		start := time.Now()
		uri := req.RequestURI
		method := req.Method

		if method == http.MethodGet {
			next.ServeHTTP(res, req)
			duration := time.Since(start)
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
				ResponseWriter: res,
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, req)
			sugar.Infoln(
				"status", responseData.status,
				"size", responseData.size,
			)
		}

		if method == http.MethodDelete {
			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: res,
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, req)
			sugar.Infoln(
				"status", responseData.status,
				"size", responseData.size,
			)
		}
	}

	return http.HandlerFunc(logFn)
}

// PanicRecoveryMiddleware восстанавливает приложение после паники.
func (con *Controller) PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				con.sugar.Errorf("Error recovering from panic: %v", err)
				http.Error(res, "Error recovering from panic", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(res, req)
	})
}

// ShortenURL обрабатывает запросы на создание сокращенного URL из входящего URL.
//
// HTTP-ответ:
//   - 401 Unauthorized: если пользователь не аутентифицирован.
//   - 201 Created: если сокращение URL завершилось успешно.
//   - 409 Conflict: если оригинальный URL уже существует в базе.
//   - 400 Bad Request: если произошла ошибка при записи ответа.
func (con *Controller) ShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var originalURL string

		if strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			originalURL = extractURLfromJSON(res, req)
		} else if strings.Contains(req.Header.Get("Content-Type"), "text/html") {
			originalURL = extractURLfromHTML(res, req)
		} else {
			b, _ := io.ReadAll(req.Body)
			originalURL = string(b)
		}

		userID := req.Header.Get("User-ID")
		if userID == "" {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		shortID, errUpdateData := con.storageService.UpdateData(req, originalURL, userID)

		con.userService.AddURLs(con.conf.BaseURL, userID, shortID, originalURL)

		if errUpdateData != nil && errors.Is(errUpdateData, repository.ErrDuplicateURL) {
			res.WriteHeader(http.StatusConflict)
		} else {
			res.WriteHeader(http.StatusCreated)
		}

		_, err := res.Write([]byte(con.conf.BaseURL + "/" + shortID))
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

// APIShortenURL предоставляет API для создания сокращенного URL из входящего JSON-запроса.
//
// HTTP-ответ:
//   - 401 Unauthorized: если пользователь не аутентифицирован.
//   - 201 Created: если сокращение URL завершилось успешно.
//   - 409 Conflict: если оригинальный URL уже существует в базе.
//   - 400 Bad Request: если произошла ошибка при записи ответа или сериализации.
func (con *Controller) APIShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		originalURL := extractURLfromJSON(res, req)
		userID := req.Header.Get("User-ID")
		if userID == "" {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		shortID, errUpdateData := con.storageService.UpdateData(req, originalURL, userID)

		con.userService.AddURLs(con.conf.BaseURL, userID, shortID, originalURL)

		shorturl.URL = con.conf.BaseURL + "/" + shortID

		resp, errMarshal := json.Marshal(shorturl)
		if errMarshal != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json")

		if errUpdateData != nil && errors.Is(errUpdateData, repository.ErrDuplicateURL) {
			res.WriteHeader(http.StatusConflict)
		} else {
			res.WriteHeader(http.StatusCreated)
		}

		_, err := res.Write(resp)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

// APIShortenBatchURL обрабатывает пакетные запросы на создание сокращенных URL из JSON-запроса.
//
// HTTP-ответ:
//   - 401 Unauthorized: если пользователь не аутентифицирован.
//   - 201 Created: если пакетное сокращение URL завершилось успешно.
//   - 409 Conflict: если один из оригинальных URL уже существует в базе.
//   - 400 Bad Request: если произошла ошибка при обработке запроса или сериализации.
func (con *Controller) APIShortenBatchURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		urls := extractURLsfromJSONBatchRequest(req)
		if urls == nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		userID := req.Header.Get("User-ID")
		if userID == "" {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}

		batchResponse := []batchResponseEntity{}
		var errUpdateData error
		for _, url := range urls {
			shortID, err := con.storageService.UpdateData(req, url.OriginalURL, userID)
			errUpdateData = err

			if err == nil {
				con.userService.AddURLs(con.conf.BaseURL, userID, shortID, url.OriginalURL)
			}

			batchResponse = append(batchResponse, batchResponseEntity{
				CorrelationID: url.CorrelationID,
				ShortURL:      con.conf.BaseURL + "/" + shortID})
		}

		resp, err := json.Marshal(batchResponse)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json")

		if errUpdateData != nil && errors.Is(errUpdateData, repository.ErrDuplicateURL) {
			res.WriteHeader(http.StatusConflict)
		} else {
			res.WriteHeader(http.StatusCreated)
		}

		_, err = res.Write(resp)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

// GetOriginalURL восстанавливает оригинальный URL из сокращенного идентификатора.
//
// HTTP-ответ:
//   - 307 Temporary Redirect: перенаправление на оригинальный URL, если он найден.
//   - 410 Gone: если URL был удален.
//   - 400 Bad Request: если была ошибка при получении данных.
func (con *Controller) GetOriginalURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/")

		originalURL, isDeleted, err := con.storageService.GetData(id)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
		if isDeleted {
			res.WriteHeader(http.StatusGone)
			http.Error(res, "Gone", http.StatusGone)
			return
		}

		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}

// PingHandler проверяет соединение с хранилищем данных.
//
// HTTP-ответ:
//   - 200 OK: если соединение успешно.
//   - 500 Internal Server Error: если соединение не удалось.
func (con *Controller) PingHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := con.storageService.Ping()
		if err != nil {
			con.sugar.Errorf("Database connection error: %v", err)
			http.Error(res, "Database connection error", http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
		con.sugar.Info("connected to the database successfully")
	}
}
