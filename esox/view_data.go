package esox

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

type ViewData struct {
	w http.ResponseWriter
	r *http.Request

	template     *template.Template
	templateName string

	flashCookieDeleted bool

	Flashes []FlashData
}

func (d *ViewData) Flash(level FlashLevel, message string) *ViewData {
	d.Flashes = append(d.Flashes, FlashData{level, message})
	return d
}

func (d *ViewData) FlashInfo(message string) *ViewData {
	return d.Flash(FlashLevelInfo, message)
}

func (d *ViewData) FlashSuccess(message string) *ViewData {
	return d.Flash(FlashLevelSuccess, message)
}

func (d *ViewData) FlashWarning(message string) *ViewData {
	return d.Flash(FlashLevelWarning, message)
}

func (d *ViewData) FlashError(message string) *ViewData {
	return d.Flash(FlashLevelError, message)
}

func (d *ViewData) setFlashCookie(redirect bool) {
	if !redirect || d.flashCookieDeleted {
		return
	}

	http.SetCookie(d.w, &http.Cookie{
		Name:     "flash",
		Value:    encodeFlashCookie(d.Flashes),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

func (d *ViewData) Redirect(url string, code int) {
	d.setFlashCookie(true)
	http.Redirect(d.w, d.r, url, code)
}

type RenderData interface {
	SetFlashes(flashes []FlashData)
}

func (d *ViewData) Render(code int, data RenderData) {
	d.setFlashCookie(false)

	log := hlog.FromRequest(d.r).With().Str("template", d.templateName).Logger()

	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesBufferPool.Put(buf)

	data.SetFlashes(d.Flashes)
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
