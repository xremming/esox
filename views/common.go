package views

import (
	"net/http"

	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/forms"
)

type data struct {
	Title       string
	Description string
	Flashes     []flash.Data
	Form        forms.Form
	Data        any
}

func (d *data) SetFlashes(flashes []flash.Data) {
	d.Flashes = flashes
}

var errorTmpl = esox.GetTemplate("error.html", "base.html")

type renderErrorData struct {
	StatusCode   int
	StatusText   string
	ErrorMessage string
	RequestID    string
}

func renderError(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
	errorData := renderErrorData{
		StatusCode:   statusCode,
		StatusText:   http.StatusText(statusCode),
		ErrorMessage: errorMessage,
	}

	requestID, ok := hlog.IDFromRequest(r)
	if ok {
		errorData.RequestID = requestID.String()
	}

	errorTmpl.Render(w, r, statusCode, &data{Data: errorData})
}
