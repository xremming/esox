package views

import (
	"net/http"

	"github.com/xremming/abborre/esox"
)

var homeTmpl = esox.GetTemplate("home.html", "base.html")

func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeTmpl.Render(w, r, 200, &data{})
	}
}

var storyTmpl = esox.GetTemplate("story.html", "base.html")

func Story() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storyTmpl.Render(w, r, 200, &data{})
	}
}
