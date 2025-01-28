package handlers

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/service"
	"shortener/internal/storage"
	"shortener/internal/user"
	"strconv"
	"time"

	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Controller struct {
	conf  *config.Config
	st    storage.Storage
	sugar *zap.SugaredLogger
	user  user.User
}

type contextKey string

const userIDKey contextKey = "userID"

func NewController(conf *config.Config, st storage.Storage, logger *zap.SugaredLogger, usr user.User) *Controller {
	return &Controller{
		conf:  conf,
		st:    st,
		sugar: logger,
		user:  usr,
	}
}

func (con *Controller) APIGetUserURLs() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userID, ok := req.Context().Value(userIDKey).(string)
		if !ok {
			return
		}

		urls, exist := con.user.GetUserURLs(userID)
		if !exist {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		if len(urls) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		response, err := json.Marshal(urls)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		_, err = res.Write(response)
		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
	}
}

func (con *Controller) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("AuthToken")
		if err != nil || cookie == nil {
			con.sugar.Debugf("(Authenticate) Missing or invalid token, generating new session token")
			userID := uuid.New().String()
			con.user.SetUserIDCookie(w, userID)

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		uid, err := con.user.GetUserIDCookie(r)
		if err != nil {

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			con.sugar.Debugf("(Authenticate) Unauthorized")
		}
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
	}

	return http.HandlerFunc(logFn)
}

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

		shortID, errUpdateData := con.st.UpdateData(originalURL)

		userID, _ := req.Context().Value(userIDKey).(string)
		con.user.AddURLs(con.conf.BaseURL, userID, shortID, originalURL)

		if errUpdateData != nil && errors.Is(errUpdateData, service.ErrDuplicateURL) {
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

func (con *Controller) APIShortenURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		originalURL := extractURLfromJSON(res, req)

		shortID, errUpdateData := con.st.UpdateData(originalURL)

		userID, _ := req.Context().Value(userIDKey).(string)
		con.user.AddURLs(con.conf.BaseURL, userID, shortID, originalURL)

		shorturl.URL = con.conf.BaseURL + "/" + shortID

		resp, errMarshal := json.Marshal(shorturl)
		if errMarshal != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json")

		if errUpdateData != nil && errors.Is(errUpdateData, service.ErrDuplicateURL) {
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

func (con *Controller) APIShortenBatchURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		urls := extractURLsfromJSONBatchRequest(req)
		if urls == nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		batchResponse := []batchResponseEntity{}
		var errUpdateData error
		for _, url := range urls {
			shortID, err := con.st.UpdateData(url.OriginalURL)
			errUpdateData = err

			userID, _ := req.Context().Value(userIDKey).(string)
			if err == nil {
				con.user.AddURLs(con.conf.BaseURL, userID, shortID, url.OriginalURL)
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

		if errUpdateData != nil && errors.Is(errUpdateData, service.ErrDuplicateURL) {
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

func (con *Controller) PingHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := con.st.Ping()
		if err != nil {
			con.sugar.Errorf("Database connection error: %v", err)
			http.Error(res, "Database connection error", http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
		con.sugar.Info("connected to the database successfully")
	}
}
