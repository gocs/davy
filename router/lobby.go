package router

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gocs/davy/models"
	"github.com/gocs/davy/servererrors"
	"github.com/gorilla/csrf"
)

type LobbyPayload struct {
	CSRF    template.HTML
	Title   string
	User    string
	Joined  bool
	Players []string
	Code    string
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

	players, err := l.GetPlayers()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	a.tmpl.ExecuteTemplate(w, "lobby.html", LobbyPayload{
		CSRF:    csrf.TemplateField(r),
		Title:   "Lobby",
		User:    username,
		Joined:  true,
		Code:    code,
		Players: players,
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
