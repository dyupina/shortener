package user

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserService(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service)
}

func TestGetUserIDFromCookie(t *testing.T) {
	service := NewUserService()
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	uid := "12345"
	err := service.SetUserIDCookie(res, uid)
	assert.NoError(t, err)

	req.Header.Set("Cookie", res.Header().Get("Set-Cookie"))

	retrievedUID, err := service.GetUserIDFromCookie(req)
	assert.NoError(t, err)
	assert.Equal(t, uid, retrievedUID)
}

func TestSetUserIDCookie(t *testing.T) {
	service := NewUserService()
	res := httptest.NewRecorder()
	uid := "12345"

	err := service.SetUserIDCookie(res, uid)
	assert.NoError(t, err)

	cookie := res.Header().Get("Set-Cookie")
	assert.Contains(t, cookie, "AuthToken")
}

func TestGetUserURLs(t *testing.T) {
	service := NewUserService()
	userID := "user123"

	urls, exist := service.GetUserURLs(userID)
	assert.False(t, exist)
	assert.Nil(t, urls)

	service.AddURLs("http://base.com", userID, "short", "http://original.com")
	urls, exist = service.GetUserURLs(userID)
	assert.True(t, exist)
	assert.Len(t, urls, 1)
	assert.Equal(t, "http://base.com/short", urls[0].ShortURL)
	assert.Equal(t, "http://original.com", urls[0].OriginalURL)
}

func TestAddURLs(t *testing.T) {
	service := NewUserService()
	userID := "user123"
	urlCount := 5

	for i := 0; i < urlCount; i++ {
		service.AddURLs("http://base.com", userID, "short"+strconv.Itoa(i), "http://original"+strconv.Itoa(i)+".com")
	}

	urls, exist := service.GetUserURLs(userID)
	assert.True(t, exist)
	assert.Len(t, urls, urlCount)
}

func TestInitUserURLs(t *testing.T) {
	service := NewUserService()
	userID := "user123"

	service.InitUserURLs(userID)
	urls, exist := service.GetUserURLs(userID)
	assert.True(t, exist)
	assert.Empty(t, urls)
}
