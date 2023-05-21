package views

import (
	"net/http"

	"github.com/xremming/abborre/esox"
)

var homeTmpl = esox.GetTemplate("home.html", "base.html")

func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeTmpl.Render(w, r, 200, &data{Title: "Home"})
	}
}
