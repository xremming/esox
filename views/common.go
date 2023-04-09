package views

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/xremming/abborre/esox"
)

var defaultNavItems = []esox.NavItem{
	{Name: "Home", URL: "/"},
	{Name: "Events", URL: "/events"},
	{Name: "Create Event", URL: "/events/create"},
}

//go:embed templates/*
var templates embed.FS

var renderer = esox.NewRenderer(templates, "templates", "base.html")

var errorTmpl = renderer.GetTemplate("error.html")

func renderError(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
	errorTmpl.ViewData(w, r, fmt.Sprintf("%v: %v", statusCode, errorMessage)).
		WithNavItems(defaultNavItems).
		WithDescription(errorMessage).
		WithData(struct {
			StatusCode   int
			StatusText   string
			ErrorMessage string
		}{statusCode, http.StatusText(statusCode), errorMessage}).
		Render(statusCode)
}
