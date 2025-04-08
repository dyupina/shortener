package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"shortener/internal/config"
	"shortener/internal/logger"
	"shortener/internal/user"
	"testing"

	"github.com/google/uuid"
)

var uid = uuid.New().String()

func prepare() *Controller {
	c := config.NewConfig()
	s := SelectStorage(c)
	sugarLogger, _ := logger.NewLogger()
	userService := user.NewUserService()
	controller := NewController(c, s, sugarLogger, userService)

	controller.userService.InitUserURLs(uid)

	return controller
}

func auth(res http.ResponseWriter, req *http.Request, controller *Controller, uid string) {
	if err := controller.userService.SetUserIDCookie(res, uid); err != nil {
		controller.sugar.Errorf("(Authenticate) Failed to set user ID cookie: %s", err.Error())
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("User-ID", uid)
}

func BenchmarkShortenURL(b *testing.B) {
	controller := prepare()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString("https://example.com"))
	w := httptest.NewRecorder()
	auth(w, r, controller, uid)
	for i := 0; i < b.N; i++ {
		handler := controller.ShortenURL()
		handler.ServeHTTP(w, r)

		// res2 := w.Result()
		// defer res2.Body.Close()
		// shortURLfromServer, _ := io.ReadAll(res2.Body)
		// fmt.Printf(">>%s\n", shortURLfromServer)
	}
}

func BenchmarkGetOriginalURL(b *testing.B) {
	controller := prepare()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString("https://example.com"))
	w := httptest.NewRecorder()
	auth(w, r, controller, uid)

	handler := controller.ShortenURL()
	handler.ServeHTTP(w, r)
	res1 := w.Result()
	defer func() {
		if err := res1.Body.Close(); err != nil {
			controller.sugar.Errorf("res1.Body.Close() error")
		}
	}()
	shortURLfromServer, _ := io.ReadAll(res1.Body)

	r2 := httptest.NewRequest("GET", string(shortURLfromServer), nil)
	w2 := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handler2 := controller.GetOriginalURL()
		handler2.ServeHTTP(w2, r2)

		// res1 := w2.Result()
		// defer res1.Body.Close()
		// resp, _ := io.ReadAll(res1.Body)
		// fmt.Printf(">%s\n", resp)
	}
}

func BenchmarkAPIShortenURL(b *testing.B) {
	controller := prepare()
	r := httptest.NewRequest("POST", "/api/shorten", bytes.NewBufferString(fmt.Sprintf(`{"url":"%s"}`, "https://example.com")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	auth(w, r, controller, uid)
	for i := 0; i < b.N; i++ {
		handler := controller.APIShortenURL()
		handler.ServeHTTP(w, r)

		// res2 := w.Result()
		// defer res2.Body.Close()
		// resp, _ := io.ReadAll(res2.Body)
		// fmt.Printf(">>%s\n", resp)
	}
}

const batch = `[
	{
		"correlation_id": "id1",
		"original_url": "http://example.com/1"
	},
	{
		"correlation_id": "id2",
		"original_url": "http://example.com/2"
	}
]`

func BenchmarkAPIShortenBatchURL(b *testing.B) {
	controller := prepare()
	r := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(batch))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	auth(w, r, controller, uid)
	for i := 0; i < b.N; i++ {
		handler := controller.APIShortenBatchURL()
		handler.ServeHTTP(w, r)

		// res1 := w.Result()
		// defer res1.Body.Close()
		// resp, _ := io.ReadAll(res1.Body)
		// fmt.Printf(">%s\n", resp)
	}
}

func BenchmarkAPIGetUserURLs(b *testing.B) {
	controller := prepare()
	r := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(batch))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	auth(w, r, controller, uid)
	handler := controller.APIShortenBatchURL()
	handler.ServeHTTP(w, r)

	// res1 := w.Result()
	// defer res1.Body.Close()
	// shortURLfromServer, _ := io.ReadAll(res1.Body)
	// fmt.Printf(">%s\n", shortURLfromServer)

	r2 := httptest.NewRequest("GET", "/api/user/urls", nil)
	w2 := httptest.NewRecorder()
	auth(w2, r2, controller, uid)
	for i := 0; i < b.N; i++ {
		handler := controller.APIGetUserURLs()
		handler.ServeHTTP(w2, r2)

		// res2 := w2.Result()
		// defer res2.Body.Close()
		// shortURLfromServer, _ := io.ReadAll(res2.Body)
		// fmt.Printf(">>%s\n", shortURLfromServer)
	}
}

func BenchmarkDeleteUserURLs(b *testing.B) {
	controller := prepare()

	r := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(batch))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	auth(w, r, controller, uid)
	handler := controller.APIShortenBatchURL()
	handler.ServeHTTP(w, r)

	res1 := w.Result()
	defer func() {
		if err := res1.Body.Close(); err != nil {
			controller.sugar.Errorf("res1.Body.Close() error")
		}
	}()
	resp, _ := io.ReadAll(res1.Body)
	var batchResp []batchResponseEntity
	err := json.Unmarshal(resp, &batchResp)
	if err != nil {
		return
	}
	urlToDel := path.Base(batchResp[0].ShortURL)
	r2 := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBufferString(fmt.Sprintf("[\"%s\"]", urlToDel)))
	r2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	auth(w2, r2, controller, uid)
	for i := 0; i < b.N; i++ {
		handler := controller.DeleteUserURLs()
		handler.ServeHTTP(w2, r2)
	}
}
