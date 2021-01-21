package loader

import (
	"html/template"
	"net/http"
)

// Templates embeds a templates for easier serving templates
type Templates struct {
	templates *template.Template
}

// NewTemplates creates and serves templates using the directory/location of the files
func NewTemplates(path string) *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob(path)),
	}
}

// ExecuteTemplate local implementation for the html execute template
func (t *Templates) ExecuteTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, tmpl, data)
}
