package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type UserURL struct {
	UUID        string `db:"user_id"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
	DeletedFlag bool   `db:"is_deleted"`
}

type user struct {
	urls       map[string][]UserURL
	cookieName string
	cookie     *securecookie.SecureCookie
}

type UserService interface {
	GetUserIDFromCookie(r *http.Request) (string, error)
	SetUserIDCookie(res http.ResponseWriter, uid string) error
	GetUserURLs(userID string) ([]UserURL, bool)
	AddURLs(baseURL, userID, shortURL, originalURL string)
	InitUserURLs(userID string)
}

func newSecurecookie() *securecookie.SecureCookie {
	var hashKey = []byte("very-very-very-very-secret-key32")
	var blockKey = []byte("a-lot-of-secret!")
	return securecookie.New(hashKey, blockKey)
}

func NewUserService() UserService {
	return &user{
		urls:       make(map[string][]UserURL),
		cookieName: "AuthToken",
		cookie:     newSecurecookie(),
	}
}

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

func (u *user) GetUserURLs(userID string) ([]UserURL, bool) {
	urls, exist := u.urls[userID]
	return urls, exist
}

const estimatedSize = 100

func (u *user) AddURLs(baseURL, userID, shortURL, originalURL string) {
	if _, exists := u.urls[userID]; !exists { // memory optimisation
		u.urls[userID] = make([]UserURL, 0, estimatedSize)
	}

	short := baseURL + "/" + shortURL
	u.urls[userID] = append(u.urls[userID], UserURL{ShortURL: short, OriginalURL: originalURL})
}

func (u *user) InitUserURLs(userID string) {
	u.urls[userID] = []UserURL{}
}
