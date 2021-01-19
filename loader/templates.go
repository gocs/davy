package loader

import (
	"html/template"
	"net/http"
)

type Templates struct {
	templates *template.Template
}

func NewTemplates(path string) *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob(path)),
	}
}

func (t *Templates) ExecuteTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, tmpl, data)
}
