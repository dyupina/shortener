package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"shortener/internal/config"
	"shortener/internal/logger"
	"shortener/internal/mocks"
	"shortener/internal/storage"
	"shortener/internal/user"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SelectStorage(c *config.Config) storage.StorageService {
	if c.DBConnection != "" {
		log.Printf("try using DB\n")
		s := storage.NewStorageDB(c.DBConnection)
		return s
	}

	if c.URLStorageFile != "" {
		log.Printf("try using file\n")
		s := storage.NewStorageFile(c)
		if s != nil {
			err := storage.RestoreURLstorage(c, s)
			if err != nil {
				log.Printf(" restore error\n")
			} else {
				storage.AutoSave(s)
				return s
			}
		} else {
			log.Printf(" error using file")
		}
	}

	log.Printf("using memory\n")
	s := storage.NewStorageMemory()

	return s
}

func TestAPIShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		data         string
		expectedCode int
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example.com"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example123123.com"},
	}
	c := config.NewConfig()
	s := SelectStorage(c)
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/api/shorten", bytes.NewBufferString(fmt.Sprintf(`{"url":"%s"}`, tc.data)))
			w := httptest.NewRecorder()

			handler := controller.Authenticate(controller.ShortenURL())
			handler.ServeHTTP(w, r)

			res := w.Result()

			require.Equal(t, tc.expectedCode, res.StatusCode, "Response code does not match expected")
			defer func() {
				if err := res.Body.Close(); err != nil {
					controller.sugar.Errorf("res.Body.Close() error")
				}
			}()
		})
	}
}

func TestShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		data         string
		expectedCode int
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example.com"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example123123.com"},
	}

	c := config.NewConfig()
	s := SelectStorage(c)
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBufferString(tc.data))
			w := httptest.NewRecorder()

			handler := controller.Authenticate(controller.ShortenURL())
			handler.ServeHTTP(w, r)

			res := w.Result()

			require.Equal(t, tc.expectedCode, res.StatusCode, "Response code does not match expected")
			defer func() {
				if err := res.Body.Close(); err != nil {
					controller.sugar.Errorf("res.Body.Close() error")
				}
			}()
		})
	}
}

func prepare_(t *testing.T) (*mocks.MockStorageService, *mocks.MockUserService, *Controller) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sugarLogger, _ := logger.NewLogger()
	conf := config.NewConfig()
	// _ = config.Init(conf) // TODO ???
	mockStorageService := mocks.NewMockStorageService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	controller := NewController(conf, mockStorageService, sugarLogger, mockUserService)

	return mockStorageService, mockUserService, controller
}

func TestGetOriginalURL(t *testing.T) {
	tests := []struct {
		mockSetup        func(storSrv *mocks.MockStorageService, controller *Controller)
		name             string
		requestPath      string
		expectedLocation string
		expectedStatus   int
	}{
		{
			name:        "GetOriginalURL ok",
			requestPath: "/url1",
			mockSetup: func(storSrv *mocks.MockStorageService, controller *Controller) {
				storSrv.EXPECT().GetData("url1").Return("http://example.com/1", false, nil)
			},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "http://example.com/1",
		},
		{
			name:        "GetOriginalURL url not found",
			requestPath: "/notfound",
			mockSetup: func(storSrv *mocks.MockStorageService, controller *Controller) {
				storSrv.EXPECT().GetData("notfound").Return("", false, errors.New("not found"))
			},
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:        "GetOriginalURL URL isDeleted",
			requestPath: "/isDeleted",
			mockSetup: func(storSrv *mocks.MockStorageService, controller *Controller) {
				storSrv.EXPECT().GetData("isDeleted").Return("", true, nil)
			},
			expectedStatus:   http.StatusGone,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storSrv, _, controller := prepare_(t)
			tt.mockSetup(storSrv, controller)

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()

			handler := controller.GetOriginalURL()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedLocation != "" {
				location, err := resp.Location()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLocation, location.String())
			}

			if err := resp.Body.Close(); err != nil {
				controller.sugar.Errorf("resp.Body.Close() error")
			}
		})
	}
}
func TestDeleteUserURLs(t *testing.T) {
	tests := []struct {
		mockSetup      func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request)
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:        "DeleteUserURLs ok",
			requestBody: "[\"url1\"]",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
				uid := "testUserID"
				userSrv.EXPECT().SetUserIDCookie(w, uid).Return(nil)
				req.Header.Set("User-ID", uid)

				storSrv.EXPECT().BatchDeleteURLs(uid, gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "DeleteUserURLs Unauthorized",
			requestBody: "", // не важно
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
				// do nothing
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "DeleteUserURLs Bad Request",
			requestBody: "", // invalid
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
				uid := "testUserID"
				userSrv.EXPECT().SetUserIDCookie(w, uid).Return(nil)
				req.Header.Set("User-ID", uid)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storSrv, userSrv, controller := prepare_(t)

			req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			tt.mockSetup(storSrv, userSrv, w, req)

			handler := controller.DeleteUserURLs()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if err := resp.Body.Close(); err != nil {
				controller.sugar.Errorf("resp.Body.Close() error")
			}
		})
	}
}

func TestAPIGetUserURLs(t *testing.T) {
	tests := []struct {
		mockSetup      func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request)
		name           string
		expectedStatus int
	}{
		{
			name: "APIGetUserURLs ok",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
				uid := "testUserID"
				userSrv.EXPECT().SetUserIDCookie(w, uid).Return(nil)
				req.Header.Set("User-ID", uid)

				userSrv.EXPECT().GetUserURLs(uid).Return([]user.UserURL{
					{ShortURL: "url1", OriginalURL: "http://example.com/1"},
					{ShortURL: "url2", OriginalURL: "http://example.com/2"},
				}, true)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "APIGetUserURLs StatusUnauthorized",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "APIGetUserURLs StatusUnauthorized (URL doesn't exists)",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
				uid := "testUserID"
				userSrv.EXPECT().SetUserIDCookie(w, uid).Return(nil)
				req.Header.Set("User-ID", uid)
				userSrv.EXPECT().GetUserURLs(uid).Return([]user.UserURL{
					{ShortURL: "url1", OriginalURL: "http://example.com/1"},
					{ShortURL: "url2", OriginalURL: "http://example.com/2"},
				}, false)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "APIGetUserURLs StatusNoContent",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request) {
				uid := "testUserID"
				userSrv.EXPECT().SetUserIDCookie(w, uid).Return(nil)
				req.Header.Set("User-ID", uid)
				userSrv.EXPECT().GetUserURLs(uid).Return([]user.UserURL{}, true)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storSrv, userSrv, controller := prepare_(t)

			req := httptest.NewRequest("GET", "/api/user/urls", nil)
			w := httptest.NewRecorder()

			tt.mockSetup(storSrv, userSrv, w, req)

			handler := controller.APIGetUserURLs()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if err := resp.Body.Close(); err != nil {
				controller.sugar.Errorf("resp.Body.Close() error")
			}
		})
	}
}

func TestAPIShortenBatchURL(t *testing.T) {
	tests := []struct {
		mockSetup      func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request, controller *Controller)
		requestBody    interface{}
		name           string
		expectedBody   []batchResponseEntity
		expectedStatus int
	}{
		{
			name: "APIShortenBatchURL ok",
			requestBody: []batchRequestEntity{
				{
					CorrelationID: "id1",
					OriginalURL:   "http://example.com/1",
				},
			},
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request, controller *Controller) {
				uid := "testUserID"
				userSrv.EXPECT().SetUserIDCookie(w, uid).Return(nil)
				req.Header.Set("User-ID", uid)

				storSrv.EXPECT().UpdateData(req, "http://example.com/1", uid).Return("url1", nil)
				userSrv.EXPECT().AddURLs(controller.conf.BaseURL, uid, "url1", "http://example.com/1")
			},
			expectedStatus: http.StatusCreated,
			expectedBody: []batchResponseEntity{
				{
					CorrelationID: "id1",
					ShortURL:      "http://localhost:8080/url1",
				},
			},
		},
		{
			name: "APIShortenBatchURL StatusUnauthorized",
			requestBody: []batchRequestEntity{
				{
					CorrelationID: "id1",
					OriginalURL:   "http://example.com/1",
				},
			},
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request, controller *Controller) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storSrv, userSrv, controller := prepare_(t)
			var req *http.Request
			var w *httptest.ResponseRecorder

			if tt.requestBody != nil {
				reqBody, _ := json.Marshal(tt.requestBody)

				req = httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(string(reqBody)))
				w = httptest.NewRecorder()
			}

			tt.mockSetup(storSrv, userSrv, w, req, controller)

			handler := controller.APIShortenBatchURL()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if resp.StatusCode == http.StatusCreated {
				var responseBody []batchResponseEntity
				err := json.NewDecoder(resp.Body).Decode(&responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, responseBody)
			}

			if err := resp.Body.Close(); err != nil {
				controller.sugar.Errorf("resp.Body.Close() error")
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		mockSetup      func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request, controller *Controller)
		name           string
		expectedStatus int
	}{
		{
			name: "PingHandler ok",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request, controller *Controller) {
				storSrv.EXPECT().Ping().Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "PingHandler ok",
			mockSetup: func(storSrv *mocks.MockStorageService, userSrv *mocks.MockUserService, w *httptest.ResponseRecorder, req *http.Request, controller *Controller) {
				storSrv.EXPECT().Ping().Return(errors.New(""))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storSrv, userSrv, controller := prepare_(t)
			req := httptest.NewRequest("GET", "/ping", nil)
			w := httptest.NewRecorder()
			tt.mockSetup(storSrv, userSrv, w, req, controller)

			handler := controller.PingHandler()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if err := resp.Body.Close(); err != nil {
				controller.sugar.Errorf("resp.Body.Close() error")
			}
		})
	}
}
