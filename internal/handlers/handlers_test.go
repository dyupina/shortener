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
	controller := NewController(c, s)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/api/shorten", bytes.NewBufferString(fmt.Sprintf(`{"url":"%s"}`, tc.data)))
			w := httptest.NewRecorder()

			handler := controller.ShortenURL()
			handler.ServeHTTP(w, r)

			res := w.Result()
			// проверяем код ответа
			require.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
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
	controller := NewController(c, s)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBufferString(tc.data))
			w := httptest.NewRecorder()

			handler := controller.ShortenURL()
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
		orig         string
		expectedCode int
	}{
		{method: http.MethodGet, orig: "https://example_1.com", expectedCode: http.StatusTemporaryRedirect},
		{method: http.MethodGet, orig: "https://example_2.com", expectedCode: http.StatusTemporaryRedirect},
	}

	c := config.NewConfig()
	s := storage.NewURLstorage()
	controller := NewController(c, s)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			// Отправить запрос на сокращение ссылки tc.orig
			r := httptest.NewRequest("POST", c.BaseURL, bytes.NewBufferString(tc.orig))
			w := httptest.NewRecorder()

			handler := controller.ShortenURL()
			handler.ServeHTTP(w, r)
			res1 := w.Result()
			defer res1.Body.Close()

			// Получить сокращённую ссылку из ответа
			shortURLfromServer, _ := io.ReadAll(res1.Body)

			// Отправить GET-запрос для получения исходной ссылки по краткой
			r2 := httptest.NewRequest(tc.method, string(shortURLfromServer), nil)
			w2 := httptest.NewRecorder()

			handler2 := controller.GetOriginalURL()
			handler2.ServeHTTP(w2, r2)

			res2 := w2.Result()
			defer res2.Body.Close()

			respGetBody, _ := io.ReadAll(res2.Body)

			re := regexp.MustCompile(`href="([^"]*)"`)
			match := re.FindStringSubmatch(string(respGetBody))

			require.Greater(t, len(match), 1, "Ответ должен содержать исходную ссылку") // ответ должен содержать href="https://example_1.com"
			require.Equal(t, tc.orig, match[1], "Ссылки должны совпадать")

			// Проверяем код ответа
			require.Equal(t, tc.expectedCode, res2.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
