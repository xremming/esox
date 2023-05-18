package esox

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox/flash"
)

type ViewData struct {
	w http.ResponseWriter
	r *http.Request

	template     *template.Template
	templateName string

	flashCookieDeleted bool
}

func (d *ViewData) setFlashCookie(redirect bool) {
	if !redirect || d.flashCookieDeleted {
		return
	}

	flashes := flash.FromContext(d.r.Context())
	if len(flashes) == 0 {
		return
	}

	http.SetCookie(d.w, &http.Cookie{
		Name:     "flash",
		Value:    flash.Encode(flashes),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

func (d *ViewData) Redirect(url string, code int) {
	d.setFlashCookie(true)
	http.Redirect(d.w, d.r, url, code)
}

type RenderData interface {
	SetFlashes(flashes []flash.Data)
}

func (d *ViewData) Render(code int, data RenderData) {
	d.setFlashCookie(false)

	log := hlog.FromRequest(d.r).With().Str("template", d.templateName).Logger()

	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesBufferPool.Put(buf)

	flashes := flash.FromRequest(d.r)
	data.SetFlashes(flashes)
	log.Info().Interface("data", data).Msg("rendering template")
	err := d.template.Execute(buf, data)
	if err != nil {
		log.Err(err).Msg("failed to execute template")
		http.Error(
			d.w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	d.w.WriteHeader(code)
	_, err = buf.WriteTo(d.w)
	if err != nil {
		log.Err(err).Msg("failed to write buffer to response writer")
		return
	}
}
