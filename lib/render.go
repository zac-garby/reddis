package lib

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var indexTemplate *template.Template

// RenderIndex renders the index page with the given post tree.
func RenderIndex(w http.ResponseWriter, tree Node) error {
	return indexTemplate.Execute(w, tree)
}

func init() {
	loadTemplate("index", &indexTemplate)
}

func loadTemplate(name string, tmpl **template.Template) {
	t, err := template.New(name).ParseFiles(
		filepath.Join("static", name+".html"),
		filepath.Join("static", "nav.html"),
		filepath.Join("static", "head.html"),
	)

	if err != nil {
		panic(err)
	}

	(*tmpl) = t
}
