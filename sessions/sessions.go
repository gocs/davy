package sessions

import "github.com/gorilla/sessions"

// Session is used for storing session cookies
type Session struct {
	Store *sessions.CookieStore
}

// New creates new session using a secret key
func New(secret string) *Session {
	return &Session{
		Store: sessions.NewCookieStore([]byte(secret)),
	}
}
