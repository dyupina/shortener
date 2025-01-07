package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"shortener/internal/config"
	"shortener/internal/logger"
	"shortener/internal/storage"

	"github.com/stretchr/testify/require"
)

func TestAPIShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		data         string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example.com"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example123123.com"},
	}

	c := config.NewConfig()
	s := storage.NewURLstorage()
	sugarLogger, _ := logger.NewLogger()
	controller := NewController(c, s, sugarLogger)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/api/shorten", bytes.NewBufferString(fmt.Sprintf(`{"url":"%s"}`, tc.data)))
			w := httptest.NewRecorder()

			handler := controller.ShortenURL()
			handler.ServeHTTP(w, r)

			res := w.Result()

			require.Equal(t, tc.expectedCode, res.StatusCode, "Response code does not match expected")
			defer res.Body.Close()
		})
	}
}

func TestShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		data         string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example.com"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example123123.com"},
	}

	c := config.NewConfig()
	s := storage.NewURLstorage()
	sugarLogger, _ := logger.NewLogger()
	controller := NewController(c, s, sugarLogger)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBufferString(tc.data))
			w := httptest.NewRecorder()

			handler := controller.ShortenURL()
			handler.ServeHTTP(w, r)

			res := w.Result()

			require.Equal(t, tc.expectedCode, res.StatusCode, "Response code does not match expected")
			defer res.Body.Close()
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	testCases := []struct {
		method       string
		orig         string
		contentType  string
		expectedCode int
	}{
		{method: http.MethodGet, orig: "https://example_1.com", contentType: "text/plain", expectedCode: http.StatusTemporaryRedirect},
		{method: http.MethodGet, orig: "https://example_2.com", contentType: "text/plain", expectedCode: http.StatusTemporaryRedirect},
	}

	c := config.NewConfig()
	s := storage.NewURLstorage()
	sugarLogger, _ := logger.NewLogger()
	controller := NewController(c, s, sugarLogger)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest("POST", c.BaseURL, bytes.NewBufferString(tc.orig))
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()

			handler := controller.ShortenURL()
			handler.ServeHTTP(w, r)
			res1 := w.Result()
			defer res1.Body.Close()

			shortURLfromServer, _ := io.ReadAll(res1.Body)

			r2 := httptest.NewRequest(tc.method, string(shortURLfromServer), nil)
			w2 := httptest.NewRecorder()

			handler2 := controller.GetOriginalURL()
			handler2.ServeHTTP(w2, r2)

			res2 := w2.Result()
			defer res2.Body.Close()

			respGetBody, _ := io.ReadAll(res2.Body)

			re := regexp.MustCompile(`href=['"]([^'"]+)['"]`)
			match := re.FindStringSubmatch(string(respGetBody))

			require.Greater(t, len(match), 1, "The response must contain the original URL")
			require.Equal(t, tc.orig, match[1], "URLs must match")

			require.Equal(t, tc.expectedCode, res2.StatusCode, "Response code does not match expected")
		})
	}
}
