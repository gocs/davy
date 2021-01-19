package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// StoreGetter makes sure AuthRequired only uses session store's Get
type StoreGetter interface {
	Get(r *http.Request, sessionName string) (*sessions.Session, error)
}

// AuthRequired middleware that checks if the user is loggedin
func AuthRequired(store StoreGetter) func(handler http.HandlerFunc) http.HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "session")
			u, ok := session.Values["user_id"]
			// if user doesn't existed and user is nil
			if !ok || u == nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			handler.ServeHTTP(w, r)
		}
	}
}
