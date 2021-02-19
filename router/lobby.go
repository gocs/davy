package router

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gocs/davy/models"
	"github.com/gocs/davy/servererrors"
	"github.com/gorilla/csrf"
	"gopkg.in/olahol/melody.v1"
)

// LobbyPayload is the data to pass to the template
type LobbyPayload struct {
	CSRF   template.HTML
	Title  string
	User   string
	Joined bool
	Code   string
}

func (a *App) lobbyGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.sessions.Store.Get(r, "session")
	u := session.Values["user_id"]
	userID, ok := u.(int64)
	if !ok {
		servererrors.InternalServerError(w, "userID is not int64")
		return
	}

	user, err := models.GetUserByUserID(userID)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	username, err := user.GetUsername()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	l, err := user.GetLobby()
	if err != nil {
		if err == models.ErrUserNotInLobby {
			a.tmpl.ExecuteTemplate(w, "lobby.html", LobbyPayload{
				CSRF:   csrf.TemplateField(r),
				Title:  "Lobby",
				User:   username,
				Joined: false,
			})
			return
		}
		servererrors.InternalServerError(w, fmt.Sprintf("GetLobby: %v", err))
		return
	}

	code, err := l.GetCode()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	a.tmpl.ExecuteTemplate(w, "lobby.html", LobbyPayload{
		CSRF:   csrf.TemplateField(r),
		Title:  "Lobby",
		User:   username,
		Joined: true,
		Code:   code,
	})
}

func (a *App) lobbyPostHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.sessions.Store.Get(r, "session")
	u := session.Values["user_id"]
	userID, ok := u.(int64)
	if !ok {
		servererrors.InternalServerError(w, "userID is not int64")
		return
	}

	r.ParseForm()
	choice := r.PostForm.Get("choice")
	code := r.PostForm.Get("code")
	if err := models.JoinOrCreateLobby(choice, code, userID); err != nil {
		servererrors.InternalServerError(w, fmt.Sprintf("JoinOrCreateLobby: %v", err))
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func (a *App) kickPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	u, err := models.GetUserByUsername(username)
	if err != nil {
		servererrors.InternalServerError(w, fmt.Sprintf("GetUserByUsername: %v", err))
		return
	}

	l, err := u.GetLobby()
	if err != nil {
		servererrors.InternalServerError(w, fmt.Sprintf("GetLobby: %v", err))
		return
	}

	userID := u.GetUserID()
	if err := l.LeaveLobby(userID); err != nil {
		servererrors.InternalServerError(w, fmt.Sprintf("LeaveLobby: %v", err))
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func (a *App) leavePostHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.sessions.Store.Get(r, "session")
	u := session.Values["user_id"]
	userID, ok := u.(int64)
	if !ok {
		servererrors.InternalServerError(w, "userID is not int64")
		return
	}

	l, err := models.GetLobbyByUserID(userID)
	if err != nil {
		servererrors.InternalServerError(w, fmt.Sprintf("GetLobbyByUserID: %v", err))
		return
	}

	if err := l.LeaveLobby(userID); err != nil {
		servererrors.InternalServerError(w, fmt.Sprintf("LeaveLobby: %v", err))
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func (a *App) lobbyWS() http.HandlerFunc {
	lock := new(sync.Mutex)

	h := func(s *melody.Session) {
		lock.Lock()
		defer lock.Unlock()

		session, _ := a.sessions.Store.Get(s.Request, "session")
		u := session.Values["user_id"]
		userID, ok := u.(int64)
		if !ok {
			log.Println("userID is not int64")
			return
		}

		l, err := models.GetLobbyByUserID(userID)
		if err != nil {
			return
		}

		ps, err := l.GetPlayers()
		if err != nil {
			return
		}

		jstr := strings.Join(ps, " ")
		s.Write([]byte(jstr))
	}

	a.m.HandleConnect(h)
	a.m.HandleDisconnect(h)

	a.m.HandleMessage(func(s *melody.Session, b []byte) {
		<-time.After(100 * time.Millisecond)

		lock.Lock()
		defer lock.Unlock()
		a.m.BroadcastOthers(b, s)
	})

	return func(w http.ResponseWriter, r *http.Request) {
		a.m.HandleRequest(w, r)
	}
}
