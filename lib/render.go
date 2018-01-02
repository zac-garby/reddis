package lib

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	indexTemplate    *template.Template
	registerTemplate *template.Template
	userTemplate     *template.Template
)

// RenderIndex renders the index page with the given post tree.
func RenderIndex(w http.ResponseWriter, tree Node) error {
	return indexTemplate.Execute(w, tree)
}

// RenderRegister renders the register page.
func RenderRegister(w http.ResponseWriter) error {
	return registerTemplate.Execute(w, nil)
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

// RenderUser renders a user's page.
func RenderUser(w http.ResponseWriter, user *User) error {
	return userTemplate.Execute(w, user)
}

func init() {
	loadTemplate("index", &indexTemplate)
	loadTemplate("register", &registerTemplate)
	loadTemplate("user", &userTemplate)
}

func loadTemplate(name string, tmpl **template.Template) {
	t, err := template.New(name).ParseFiles(
		filepath.Join("static", fmt.Sprintf("%s/%s.html", name, name)),
		filepath.Join("static", "nav.html"),
		filepath.Join("static", "head.html"),
	)

	if err != nil {
		panic(err)
	}

	(*tmpl) = t
}
