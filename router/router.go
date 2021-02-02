package router

import (
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gocs/davy/loader"
	"github.com/gocs/davy/middleware"
	"github.com/gocs/davy/models"
	"github.com/gocs/davy/servererrors"
	"github.com/gocs/davy/sessions"
	"github.com/gorilla/mux"
)

// NewRouter creates a new router to access some pages
func NewRouter(sessionKey string) (*mux.Router, error) {
	r := mux.NewRouter()

	err := models.NewRedisDB() // shiiiiiiiiiiiiiiiiiiiiiiiiiitttttt
	if err != nil {
		return nil, err
	}
	a := App{
		sessions: sessions.New(sessionKey),
		tmpl:     loader.NewTemplates("templates/*.html"),
	}

	mar := middleware.AuthRequired(a.sessions.Store)

	r.HandleFunc("/", mar(a.indexGetHandler)).Methods("GET")
	r.HandleFunc("/", mar(a.indexPostHandler)).Methods("POST")
	r.HandleFunc("/login", a.loginGetHandler).Methods("GET")
	r.HandleFunc("/login", a.loginPostHandler).Methods("POST")
	r.HandleFunc("/logout", mar(a.logoutPostHandler)).Methods("POST")
	r.HandleFunc("/register", a.registerGetHandler).Methods("GET")
	r.HandleFunc("/register", a.registerPostHandler).Methods("POST")

	r.HandleFunc("/exam", a.examGetHandler).Methods("GET")
	r.HandleFunc("/exam", a.examPostHandler).Methods("POST")

	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.HandleFunc("/{username}", mar(a.userGetHandler)).Methods("GET")

	return r, nil
}

// App handles the state of the application
type App struct {
	sessions *sessions.Session
	tmpl     *loader.Templates
}

// IndexPayload is the data to pass to the template
type IndexPayload struct {
	Title       string
	User        string
	Updates     []*models.Update
	DisplayForm bool
}

func (a *App) indexGetHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.sessions.Store.Get(r, "session")
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
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
		switch err {
		case redis.Nil:
			session, err := a.sessions.Store.Get(r, "session")
			if err != nil {
				servererrors.InternalServerError(w, err.Error())
				return
			}
			delete(session.Values, "user_id")
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
			return
		}
		servererrors.InternalServerError(w, err.Error())
		return
	}

	updates, err := models.GetAllUpdates()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
	a.tmpl.ExecuteTemplate(w, "index.html", IndexPayload{
		Title:       "All Updates",
		User:        username,
		Updates:     updates,
		DisplayForm: true,
	})
}

func (a *App) indexPostHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.sessions.Store.Get(r, "session")
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
	u := session.Values["user_id"]
	userID, ok := u.(int64)
	if !ok {
		servererrors.InternalServerError(w, "userID is not int64")
		return
	}

	r.ParseForm()
	body := r.PostForm.Get("update")
	err = models.PostUpdate(userID, body)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *App) userGetHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.sessions.Store.Get(r, "session")
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
	u := session.Values["user_id"]
	sessionUserID, ok := u.(int64)
	if !ok {
		servererrors.InternalServerError(w, "userID is not int64")
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	if username == "favicon.ico" || username == "serviceworker.js" {
		a.tmpl.ExecuteTemplate(w, "login.html", "unknown user")
		return
	}

	user, err := models.GetUserByUsername(username)
	if err != nil {
		switch err {
		case models.ErrUserNotFound:
			a.tmpl.ExecuteTemplate(w, "login.html", "unknown user")
		default:
			servererrors.InternalServerError(w, err.Error())
		}
		return
	}

	userID := user.GetUserID()

	updates, err := models.GetUpdates(userID)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
	a.tmpl.ExecuteTemplate(w, "index.html", IndexPayload{
		Title:       username,
		Updates:     updates,
		DisplayForm: sessionUserID == userID,
	})
}

func (a *App) loginGetHandler(w http.ResponseWriter, r *http.Request) {
	a.tmpl.ExecuteTemplate(w, "login.html", nil)
}

func (a *App) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	user, err := models.AuthenticateUser(username, password)
	if err != nil {
		switch err {
		case models.ErrUserNotFound:
			a.tmpl.ExecuteTemplate(w, "login.html", LoginPayload{Error: "unknown user"})
		case models.ErrInvalidLogin:
			a.tmpl.ExecuteTemplate(w, "login.html", LoginPayload{Error: "invalid login"})
		default:
			servererrors.InternalServerError(w, err.Error())
		}
		return
	}

	userID := user.GetUserID()

	session, err := a.sessions.Store.Get(r, "session")
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
	session.Values["user_id"] = userID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *App) logoutPostHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.sessions.Store.Get(r, "session")
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}
	delete(session.Values, "user_id")
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusFound)
}

type LoginPayload struct {
	Error string
}

func (a *App) registerGetHandler(w http.ResponseWriter, r *http.Request) {
	a.tmpl.ExecuteTemplate(w, "register.html", LoginPayload{})
}

func (a *App) registerPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	err := models.RegisterUser(username, password)
	if err != nil {
		switch err {
		case models.ErrUsernameTaken:
			a.tmpl.ExecuteTemplate(w, "register.html", LoginPayload{Error: "username taken"})
		default:
			servererrors.InternalServerError(w, err.Error())
		}
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}
