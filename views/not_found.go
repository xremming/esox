package views

import (
	"net/http"
)

var notFoundTmpl = renderer.GetTemplate("404.html", "base.html")

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notFoundTmpl.Render(w, r, 404, &data{Title: "Not Found", Nav: defaultNavItems})
	}
}
