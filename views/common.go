package views

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/forms"
)

type data struct {
	LastModified time.Time
	Public       bool
	MaxAge       time.Duration

	Flashes []flash.Data
	Form    forms.Form
	Data    any
}

func (d *data) ModTime() time.Time {
	return d.LastModified
}

func (d *data) CacheControl() (bool, time.Duration) {
	return d.Public, d.MaxAge
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
