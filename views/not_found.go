package views

import (
	"net/http"

	"github.com/xremming/abborre/esox"
)

var notFoundTmpl = esox.GetTemplate("404.html", "base.html")

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notFoundTmpl.Render(w, r, 404, &data{Title: "404 - Not Found"})
	}
}
