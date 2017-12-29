package lib

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

var indexTemplate *template.Template

// RenderIndex renders the index page with the given post tree.
func RenderIndex(w http.ResponseWriter, tree Node) error {
	return indexTemplate.Execute(w, tree)
}

// RenderPosts renders the given post tree.
func RenderPosts(w http.ResponseWriter, tree Node) error {
	for _, tmpl := range indexTemplate.Templates() {
		if tmpl.Name() == "post" {
			return tmpl.Execute(w, tree)
		}
	}

	return errors.New("render: couldn't find the post template")
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
