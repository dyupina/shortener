package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
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

	controller := &Controller{}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBufferString(tc.data))
			w := httptest.NewRecorder()

			handler := controller.ShortenURL(c, s)
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

	c := *config.NewConfig()

	// Отключить автоматические редиректы (иначе была ошибка при выполнении Get запроса)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			// Отправить запрос на сокращение ссылки tc.orig
			respPost, err := client.Post(c.BaseURL, "text/plain", bytes.NewBufferString(tc.orig))
			require.NoError(t, err, "Не удалось отправить POST-запрос")
			defer respPost.Body.Close()

			// Получить сокращённую ссылку из ответа
			shortURLfromServer, _ := io.ReadAll(respPost.Body)

			// Отправить GET-запрос для получения исходной ссылки по краткой
			respGet, err := client.Get(string(shortURLfromServer))
			require.NoError(t, err, "Не удалось отправить GET-запрос")
			defer respGet.Body.Close()
			respGetBody, _ := io.ReadAll(respGet.Body)

			re := regexp.MustCompile(`href="([^"]*)"`)
			match := re.FindStringSubmatch(string(respGetBody))

			require.Greater(t, len(match), 1, "Ответ должен содержать исходную ссылку") // ответ должен содержать href="https://example_1.com"
			require.Equal(t, tc.orig, match[1], "Ссылки должны совпадать")

			// Проверяем код ответа
			require.Equal(t, tc.expectedCode, respGet.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
