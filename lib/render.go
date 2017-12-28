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
	var err error

	indexTemplate, err = template.New("index").ParseFiles(
		filepath.Join("static", "index.html"),
	)

	if err != nil {
		panic(err)
	}
}
