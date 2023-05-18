package views

import (
	"net/http"
	"os"

	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/forms"
)

type navItem struct {
	Name string
	URL  string
}

var defaultNavItems = []navItem{
	{Name: "Home", URL: "/"},
	{Name: "Events", URL: "/events"},
	{Name: "Create Event", URL: "/events/create"},
}

type data struct {
	Title       string
	Description string
	Flashes     []flash.Data
	Nav         []navItem
	Form        forms.Form
	Data        any
}

func (d *data) SetFlashes(flashes []flash.Data) {
	d.Flashes = flashes
}

var (
	templates = os.DirFS(".")
	renderer  = esox.NewRenderer(templates, "templates", "base.html")
)

var errorTmpl = renderer.GetTemplate("error.html")

type renderErrorData struct {
	StatusCode   int
	StatusText   string
	ErrorMessage string
}

func renderError(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
	errorTmpl.ViewData(w, r).
		Render(statusCode, &data{
			Data: renderErrorData{statusCode, http.StatusText(statusCode), errorMessage},
		})
}
