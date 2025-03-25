package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

// UserURL - структура для хранения информации об URL-пользователя.
type UserURL struct {
	UUID        string `db:"user_id"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
	DeletedFlag bool   `db:"is_deleted"`
}

// user реализует сервис работы с URL пользователя, включая управление куки и хранение URL.
type user struct {
	urls       map[string][]UserURL
	cookieName string
	cookie     *securecookie.SecureCookie
}

// UserService - интерфейс для управления URL пользователя и куки.
type UserService interface {
	// GetUserIDFromCookie получает идентификатор пользователя из куки.
	GetUserIDFromCookie(r *http.Request) (string, error)
	// SetUserIDCookie устанавливает куки с идентификатором пользователя.
	SetUserIDCookie(res http.ResponseWriter, uid string) error
	// GetUserURLs возвращает все URL, связанные с пользователем.
	GetUserURLs(userID string) ([]UserURL, bool)
	// AddURLs добавляет URL для пользователя.
	AddURLs(baseURL, userID, shortURL, originalURL string)
	// InitUserURLs инициализирует структуру URL для пользователя.
	InitUserURLs(userID string)
}

// newSecurecookie создаёт и возвращает новый экземпляр securecookie для кодирования и декодирования значений куки.
func newSecurecookie() *securecookie.SecureCookie {
	var hashKey = []byte("very-very-very-very-secret-key32")
	var blockKey = []byte("a-lot-of-secret!")
	return securecookie.New(hashKey, blockKey)
}

// NewUserService создаёт и возвращает новый экземпляр сервиса UserService.
func NewUserService() UserService {
	return &user{
		urls:       make(map[string][]UserURL),
		cookieName: "AuthToken",
		cookie:     newSecurecookie(),
	}
}

// GetUserIDFromCookie возвращает идентификатор пользователя из HTTP-запроса.
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

// SetUserIDCookie устанавливает HTTP-куки с идентификатором пользователя.
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

// GetUserURLs возвращает URL, принадлежащие пользователю, и флаг существования.
func (u *user) GetUserURLs(userID string) ([]UserURL, bool) {
	urls, exist := u.urls[userID]
	return urls, exist
}

const estimatedSize = 100

// AddURLs добавляет новый URL в хранилище URL пользователя.
func (u *user) AddURLs(baseURL, userID, shortURL, originalURL string) {
	if _, exists := u.urls[userID]; !exists { // memory optimisation
		u.urls[userID] = make([]UserURL, 0, estimatedSize)
	}

	short := baseURL + "/" + shortURL
	u.urls[userID] = append(u.urls[userID], UserURL{ShortURL: short, OriginalURL: originalURL})
}

// InitUserURLs инициализирует хранилище URL для пользователя.
func (u *user) InitUserURLs(userID string) {
	u.urls[userID] = []UserURL{}
}
