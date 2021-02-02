package router

import (
	"html/template"
	"net/http"

	"github.com/gocs/davy/models"
	"github.com/gocs/davy/servererrors"
	"github.com/gorilla/csrf"
)

// ExamPayload is the data to pass to the template
type ExamPayload struct {
	CSRF        template.HTML
	Title       string
	User        string
	Question    models.QuestionT
	Correct     bool
	Points      int64
	Explanation string
}

func (a *App) examGetHandler(w http.ResponseWriter, r *http.Request) {
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

	uq, err := models.GetUserQuestion(userID)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	question, err := uq.GetQuestion()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	p, err := uq.GetPoints()
	if err != nil {
		return
	}

	qt, err := models.GetQuestion(question)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	a.tmpl.ExecuteTemplate(w, "exam.html", ExamPayload{
		CSRF:     csrf.TemplateField(r),
		Title:    "Question",
		User:     username,
		Question: *qt,
		Points:   p,
	})

}

func (a *App) examPostHandler(w http.ResponseWriter, r *http.Request) {
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

	uq, err := models.GetUserQuestion(userID)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	q, err := uq.GetQuestion()
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	qt, err := models.GetQuestion(q)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	r.ParseForm()
	choice := r.PostForm.Get("choice")

	// TODO: set error whern choice is wrong
	result, err := models.UserConfirmAnswer(userID, choice)
	if err != nil {
		servererrors.InternalServerError(w, err.Error())
		return
	}

	p, err := uq.GetPoints()
	if err != nil {
		return
	}

	var e string
	// if result is correct update rank else give an explanation
	if result {
		if err := models.UpdateRank(userID, p); err != nil {
			servererrors.InternalServerError(w, err.Error())
			return
		}
	} else {
		e = "YOU HAVE ENTERED THE WRONG CHOICE!!"
	}

	a.tmpl.ExecuteTemplate(w, "exam.html", ExamPayload{
		CSRF:        csrf.TemplateField(r),
		Title:       "Question",
		User:        username,
		Question:    *qt,
		Correct:     result,
		Points:      p,
		Explanation: e,
	})
}
