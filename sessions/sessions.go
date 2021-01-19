package sessions

import "github.com/gorilla/sessions"

type Session struct {
	Store *sessions.CookieStore
}

func New(secret string) *Session {
	return &Session{
		Store: sessions.NewCookieStore([]byte(secret)),
	}
}
