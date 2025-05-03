// Package user provides functions for managing user URLs and cookies.
package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

// UserURL - structure for storing user URL information.
type UserURL struct {
	UUID        string `db:"user_id"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
	DeletedFlag bool   `db:"is_deleted"`
}

// user implements the service for handling user URLs, including cookie management and URL storage.
type user struct {
	urls       map[string][]UserURL
	cookie     *securecookie.SecureCookie
	cookieName string
}

// UserService - interface for managing user URLs and cookies.
type UserService interface {
	// GetUserIDFromCookie retrieves the user ID from a cookie.
	GetUserIDFromCookie(r *http.Request) (string, error)
	// SetUserIDCookie sets a cookie with the user ID.
	SetUserIDCookie(res http.ResponseWriter, uid string) error
	// GetUserURLs returns all URLs associated with the user.
	GetUserURLs(userID string) ([]UserURL, bool)
	// AddURLs adds URLs for the user.
	AddURLs(baseURL, userID, shortURL, originalURL string)
	// InitUserURLs initializes the URL structure for the user.
	InitUserURLs(userID string)
	// GetUserNumber returns number of users.
	GetUserNumber() int
	// GetURLsCount returns number of shortened URLs.
	GetURLsCount() int
}

// newSecurecookie creates and returns a new instance of securecookie for encoding and decoding cookie values.
func newSecurecookie() *securecookie.SecureCookie {
	var hashKey = []byte("very-very-very-very-secret-key32")
	var blockKey = []byte("a-lot-of-secret!")
	return securecookie.New(hashKey, blockKey)
}

// NewUserService creates and returns a new instance of the UserService.
func NewUserService() UserService {
	return &user{
		urls:       make(map[string][]UserURL),
		cookieName: "AuthToken",
		cookie:     newSecurecookie(),
	}
}

// GetUserIDFromCookie returns the user ID from an HTTP request.
func (u *user) GetUserIDFromCookie(req *http.Request) (string, error) {
	cookie, err := req.Cookie(u.cookieName)
	if err != nil {
		return "", err
	}

	var uid string
	if err := u.cookie.Decode(u.cookieName, cookie.Value, &uid); err != nil {
		return "", err
	}

	return uid, nil
}

// SetUserIDCookie sets an HTTP cookie with the user ID.
func (u *user) SetUserIDCookie(res http.ResponseWriter, uid string) error {
	encoded, err := u.cookie.Encode(u.cookieName, uid)

	if err == nil {
		cookie := &http.Cookie{
			Name:    u.cookieName,
			Value:   encoded,
			Path:    "/",
			Secure:  false,
			Expires: time.Now().Add(30 * 24 * time.Hour),
		}
		http.SetCookie(res, cookie)
	} else {
		fmt.Printf("(SetUserIDCookie) err %v\n", err)
	}

	return err
}

// GetUserURLs returns URLs belonging to the user and an existence flag.
func (u *user) GetUserURLs(userID string) ([]UserURL, bool) {
	urls, exist := u.urls[userID]
	return urls, exist
}

const estimatedSize = 100

// AddURLs adds a new URL to the user's URL storage.
func (u *user) AddURLs(baseURL, userID, shortURL, originalURL string) {
	if _, exists := u.urls[userID]; !exists { // memory optimisation
		u.urls[userID] = make([]UserURL, 0, estimatedSize)
	}

	short := baseURL + "/" + shortURL
	u.urls[userID] = append(u.urls[userID], UserURL{ShortURL: short, OriginalURL: originalURL})
}

// InitUserURLs initializes the URL storage for a user.
func (u *user) InitUserURLs(userID string) {
	u.urls[userID] = []UserURL{}
}

// GetUserNumber returns the number of users in the service.
func (u *user) GetUserNumber() int {
	return len(u.urls)
}

// GetUserNumber returns the number of shortened URLs in the service.
func (u *user) GetURLsCount() int {
	urlsCount := 0
	for _, userURLs := range u.urls {
		urlsCount += len(userURLs)
	}
	return urlsCount
}
