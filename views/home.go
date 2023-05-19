package views

import (
	"net/http"
)

var homeTmpl = renderer2.GetTemplate("home.html", "base.html")

func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeTmpl.Render(w, r, 200, &data{Title: "Home", Nav: defaultNavItems})
	}
}
