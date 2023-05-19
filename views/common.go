package views

import (
	"net/http"
	"os"
	"time"

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

var location *time.Location

func init() {
	var err error
	location, err = time.LoadLocation("Europe/Helsinki")
	if err != nil {
		panic(err)
	}
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
	renderer  = esox.NewR(templates, "templates")
)

var errorTmpl = renderer.GetTemplate("error.html", "base.html")

type renderErrorData struct {
	StatusCode   int
	StatusText   string
	ErrorMessage string
}

func renderError(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
	errorTmpl.Render(w, r, statusCode, &data{
		Data: renderErrorData{statusCode, http.StatusText(statusCode), errorMessage},
	})
}
