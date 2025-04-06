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

// Controller manages HTTP requests for URL shortening operations.
type Controller struct {
	conf           *config.Config
	storageService storage.StorageService
	sugar          *zap.SugaredLogger
	userService    user.UserService
}

// NewController creates and returns a new instance of Controller using the provided configuration,
// storage, logger, and user service components.
func NewController(conf *config.Config, storageService storage.StorageService, logger *zap.SugaredLogger, us user.UserService) *Controller {
	return &Controller{
		conf:           conf,
		storageService: storageService,
		sugar:          logger,
		userService:    us,
	}
}

// DeleteUserURLs handles HTTP requests to delete URLs belonging to a user.
// The stream of deleted URLs is processed asynchronously.
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

// APIGetUserURLs handles requests to retrieve all URLs associated with a user.
// Returns a JSON response with the user's URLs.
//
// HTTP Responses:
//   - 401 Unauthorized: if the user is not authenticated.
//   - 204 No Content: if the user has no associated URLs.
//   - 200 OK: successful retrieval of user's URLs in JSON format.
//   - 500 Internal Server Error: if the connection failed.
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

// Authenticate performs user authentication using HTTP cookies.
// Sets a new user ID if the cookies are missing or invalid.
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

// GzipDecodeMiddleware decodes the content of incoming HTTP requests encoded with gzip.
//
// HTTP Response:
//   - 400 Bad Request: if gzip content decoding fails.
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

// GzipEncodeMiddleware compresses outgoing HTTP responses using gzip.
// Compression criteria: client supports gzip (Accept-Encoding),
// the response content is JSON or HTML, and its size exceeds the minimum threshold for compression.
//
// HTTP Response:
//   - 400 Bad Request: if creating gzip.Writer fails.
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

// LoggingMiddleware logs information about HTTP requests and responses.
//
// Logs:
//   - Request method (GET, POST, DELETE).
//   - Request URI.
//   - Duration of request processing.
//   - Status and size of the response in case of POST and DELETE.
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

// PanicRecoveryMiddleware recovers the application after a panic.
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

// ShortenURL handles requests to create a shortened URL from an incoming URL.
//
// HTTP Responses:
//   - 401 Unauthorized: if the user is not authenticated.
//   - 201 Created: if the URL shortening was successful.
//   - 409 Conflict: if the original URL already exists in the database.
//   - 400 Bad Request: if there was an error writing the response.
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

// APIShortenURL provides an API for creating a shortened URL from an incoming JSON request.
//
// HTTP Responses:
//   - 401 Unauthorized: if the user is not authenticated.
//   - 201 Created: if URL shortening was successful.
//   - 409 Conflict: if the original URL already exists in the database.
//   - 400 Bad Request: if there was an error in writing the response or serialization.
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

// APIShortenBatchURL handles batch requests for creating shortened URLs from a JSON request.
//
// HTTP Responses:
//   - 401 Unauthorized: if the user is not authenticated.
//   - 201 Created: if batch URL shortening is successful.
//   - 409 Conflict: if one of the original URLs already exists in the database.
//   - 400 Bad Request: if an error occurred during request processing or serialization.
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

// GetOriginalURL restores the original URL from a shortened identifier.
//
// HTTP Responses:
//   - 307 Temporary Redirect: redirect to the original URL if it is found.
//   - 410 Gone: if the URL has been deleted.
//   - 400 Bad Request: if there was an error retrieving the data.
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

// PingHandler checks the connection to the data storage.
//
// HTTP Responses:
//   - 200 OK: if the connection is successful.
//   - 500 Internal Server Error: if the connection failed.
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
