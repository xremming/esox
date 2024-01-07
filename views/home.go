package views

import (
	"net/http"
	"time"

	"github.com/xremming/abborre/esox"
)

var homeTmpl = esox.GetTemplate("home.html", "base.html")

func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeTmpl.Render(w, r, 200, &data{Public: true, MaxAge: 1 * time.Hour})
	}
}

var storyTmpl = esox.GetTemplate("story.html", "base.html")

func Story() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storyTmpl.Render(w, r, 200, &data{Public: true, MaxAge: 1 * time.Hour})
	}
}
