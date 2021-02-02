package router

import (
	"html/template"
	"net/http"

	"github.com/gocs/davy/models"
	"github.com/gocs/davy/servererrors"
	"github.com/gorilla/csrf"
)

// RankPayload is the data to pass to the template of this page
type RankPayload struct {
	CSRF      template.HTML
	User      string
	UserRanks []models.RankT
	Error     string
}

func (a *App) listTopRank(w http.ResponseWriter, r *http.Request) {
	urT, err := models.TopRanks()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	a.tmpl.ExecuteTemplate(w, "rank.html", RankPayload{
		CSRF:      csrf.TemplateField(r),
		UserRanks: urT,
	})
}

func (a *App) getCurrentStandings(w http.ResponseWriter, r *http.Request) {
	session, _ := a.sessions.Store.Get(r, "session")
	u := session.Values["user_id"]
	userID, ok := u.(int64)
	if !ok {
		servererrors.InternalServerError(w, "userID is not int64")
		return
	}

	urT, err := models.GetCurrentStandings(userID)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	a.tmpl.ExecuteTemplate(w, "rank.html", RankPayload{
		CSRF:      csrf.TemplateField(r),
		UserRanks: urT,
	})
}
