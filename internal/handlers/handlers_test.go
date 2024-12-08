package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"shortener/internal/config"
	"shortener/internal/storage"

	"github.com/stretchr/testify/require"
)

func TestShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		data         string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example.com"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example123123.com"},
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodConnect, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodHead, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodOptions, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodTrace, expectedCode: http.StatusMethodNotAllowed},
	}

	c := *config.NewConfig()
	s := *storage.NewURLstorage()

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBufferString(tc.data))
			w := httptest.NewRecorder()

			handler := ShortenURL(c, s)

			handler.ServeHTTP(w, r)

			res := w.Result()
			// проверяем код ответа
			require.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
			defer res.Body.Close()
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	testCases := []struct {
		method       string
		target       string
		expectedCode int
	}{
		{method: http.MethodPost, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodConnect, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodHead, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodOptions, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodTrace, expectedCode: http.StatusMethodNotAllowed},
	}

	c := *config.NewConfig()
	s := *storage.NewURLstorage()

	s.UpdateData(generateShortID(), c.Addr) // ???????

	for k := range s.URLStorage {
		// Добавляем новые элементы в testCases
		testCases = append(testCases, struct {
			method       string
			target       string
			expectedCode int
		}{
			method:       http.MethodGet,
			target:       k,
			expectedCode: http.StatusTemporaryRedirect,
		})
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/"+tc.target, nil)
			w := httptest.NewRecorder()

			handler := GetOriginalURL(s)

			handler.ServeHTTP(w, r)

			res := w.Result()
			// проверяем код ответа
			require.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
			defer res.Body.Close()
		})
	}
}
