package views

import (
	"net/http"
)

var homeTmpl = renderer.GetTemplate("home.html")

func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeTmpl.ViewData(w, r, "Home").
			WithNavItems(defaultNavItems).
			Render(200)
	}
}
