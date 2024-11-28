package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TestURLStore = make(map[string]string)

func TestShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		data         string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, data: "https://example.com"},
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodConnect, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodHead, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodOptions, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodTrace, expectedCode: http.StatusMethodNotAllowed},
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBufferString(tc.data))
			w := httptest.NewRecorder()

			ShortenURL(w, r)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body) // тут либо "Only POST requests are allowed!",
			// либо "http://localhost:8080/" + shortID

			assert.NoError(t, err)

			assert.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")

			id := path.Base(string(resBody))
			TestURLStore[id] = urlStore[id]
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusTemporaryRedirect},

		{method: http.MethodPost, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodConnect, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodHead, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodOptions, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodTrace, expectedCode: http.StatusMethodNotAllowed},
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			for k := range TestURLStore {
				r := httptest.NewRequest(tc.method, "/"+k, nil)
				w := httptest.NewRecorder()

				GetOriginalURL(w, r)

				assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			}
		})
	}
}
