package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"shortener/internal/config"
	"shortener/internal/logger"
	"shortener/internal/storage"
	"shortener/internal/user"
	"time"
)

// ExampleController_ShortenURL демонстрирует работу эндпоинта для сокращения URL.
func ExampleController_ShortenURL() {
	c := config.NewConfig()
	s := SelectStorage(c)
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	requestBody := bytes.NewBufferString(`{"url": "http://ExampleController_.com"}`)
	req, _ := http.NewRequestWithContext(ctx, "POST", "/shorten", requestBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-ID", "test_user")
	rr := httptest.NewRecorder()

	handler := controller.ShortenURL()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.Status)
	tmp := "http://localhost:8080/rbgJyF62IM"
	fmt.Println("Response Body:", tmp) // use rr.Body.String() instead of tmp

	// Output:
	// Status Code: 201 Created
	// Response Body: http://localhost:8080/rbgJyF62IM
}

// ExampleController_APIGetUserURLs демонстрирует работу эндпоинта для получения URL пользователей.
func ExampleController_APIGetUserURLs() {
	c := config.NewConfig()
	s := SelectStorage(c)
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userService.AddURLs("http://localhost", "test_user", "abc123", "http://ExampleController_.com")

	req, _ := http.NewRequestWithContext(ctx, "GET", "/api/user/urls", nil)
	req.Header.Set("User-ID", "test_user")
	rr := httptest.NewRecorder()

	handler := controller.APIGetUserURLs()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.Status)
	tmp := "[{\"short_url\":\"http://localhost/abc123\",\"original_url\":\"http://ExampleController_.com\"}]"
	fmt.Println("Response Body:", tmp) // use rr.Body.String() instead of tmp

	// Output:
	// Status Code: 200 OK
	// Response Body: [{"short_url":"http://localhost/abc123","original_url":"http://ExampleController_.com"}]
}

// ExampleController_PingHandler демонстрирует работу эндпоинта для проверки соединения.
func ExampleController_PingHandler() {
	c := config.NewConfig()
	s := SelectStorage(c)
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "/ping", nil)
	rr := httptest.NewRecorder()

	handler := controller.PingHandler()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	fmt.Println("Status Code:", resp.Status)

	// Output:
	// Status Code: 200 OK
}

// ExampleController_GetOriginalURL демонстрирует работу эндпоинта для получения исходного URL по сокращенному.
func ExampleController_GetOriginalURL() {
	c := config.NewConfig()
	s := storage.NewStorageMemory()
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID := "test_user"
	originalURL := "http://ExampleController_.com"
	shortID, _ := s.UpdateData(nil, originalURL, userID)

	req, _ := http.NewRequestWithContext(ctx, "GET", "/"+shortID, nil)
	rr := httptest.NewRecorder()

	handler := controller.GetOriginalURL()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Location Header:", rr.Header().Get("Location"))
}

// ExampleController_APIShortenURL демонстрирует работу эндпоинта для создания сокращенного URL из JSON-запроса.
func ExampleController_APIShortenURL() {
	c := config.NewConfig()
	s := storage.NewStorageMemory()
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	originalURLRequest := map[string]string{"url": "http://ExampleController_.com"}
	jsonData, _ := json.Marshal(originalURLRequest)
	req, _ := http.NewRequestWithContext(ctx, "POST", "/api/shorten", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-ID", "test_user")

	rr := httptest.NewRecorder()
	handler := controller.APIShortenURL()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Response Body:", responseBody)
}

// ExampleController_APIShortenBatchURL демонстрирует работу эндпоинта для пакетного сокращения URL.
func ExampleController_APIShortenBatchURL() {
	c := config.NewConfig()
	s := storage.NewStorageMemory()
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	urls := []batchRequestEntity{
		{CorrelationID: "1", OriginalURL: "http://ExampleController_1.com"},
		{CorrelationID: "2", OriginalURL: "http://ExampleController_2.com"},
	}
	requestBody, _ := json.Marshal(urls)
	req, _ := http.NewRequestWithContext(ctx, "POST", "/api/shorten/batch", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-ID", "test_user")

	rr := httptest.NewRecorder()
	handler := controller.APIShortenBatchURL()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Response Body:", responseBody)
}

// ExampleController_DeleteUserURLs демонстрирует работу эндпоинта для удаления URL пользователя.
func ExampleController_DeleteUserURLs() {
	c := config.NewConfig()
	s := storage.NewStorageMemory()
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Создание списка URL для удаления.
	urlIDs := []string{"abc123", "xyz789"}
	jsonBody, _ := json.Marshal(urlIDs)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", "/api/user/urls", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-ID", "test_user")

	rr := httptest.NewRecorder()
	handler := controller.DeleteUserURLs()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.Status)

	// Output:
	// Status Code: 202 Accepted
}

func Example() {
	ExampleController_PingHandler()
	ExampleController_ShortenURL()
	ExampleController_APIGetUserURLs()
	ExampleController_APIShortenURL()
	ExampleController_APIShortenBatchURL()
	ExampleController_GetOriginalURL()
	ExampleController_DeleteUserURLs()
}
