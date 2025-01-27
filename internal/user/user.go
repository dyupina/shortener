package user

import (
	"net/http"
	"shortener/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"userID"`
}

type user struct {
	urls map[string][]models.Storage
}

type User interface {
	GetUserIDCookie(r *http.Request) (string, error)
	SetUserIDCookie(res http.ResponseWriter, uid string)
	GetUserURLs(userID string) ([]models.Storage, bool)
	AddURLs(baseURL, userID, shortURL, originalURL string)
}

var JwtKey = []byte("secretKey")
var tokenExp = 24 * time.Hour

func NewUser() *user {
	return &user{
		urls: make(map[string][]models.Storage),
	}
}

func (u *user) GetUserIDCookie(req *http.Request) (string, error) {
	cookie, err := req.Cookie("AuthToken")
	if err != nil {
		return "", err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims,
		func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

	if err != nil || !token.Valid {
		return "", err
	}

	return claims.UserID, nil
}

func (u *user) SetUserIDCookie(res http.ResponseWriter, uid string) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: uid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(res, &http.Cookie{
		Name:     "AuthToken",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	})
}

func (u *user) GetUserURLs(userID string) ([]models.Storage, bool) {
	urls, exist := u.urls[userID]
	return urls, exist
}

func (u *user) AddURLs(baseURL, userID, shortURL, originalURL string) {
	short := baseURL + "/" + shortURL
	u.urls[userID] = append(u.urls[userID], models.Storage{ShortURL: short, OriginalURL: originalURL})
}
