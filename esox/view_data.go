package esox

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

type NavItem struct {
	Name string
	URL  string
}

type ViewData struct {
	w http.ResponseWriter
	r *http.Request

	template     *template.Template
	templateName string

	flashCookieDeleted bool

	Title       string
	Description string
	Flashes     []FlashData
	NavItems    []NavItem
	Form        *FormData
	Data        any
}

func (d *ViewData) WithDescription(description string) *ViewData {
	d.Description = description
	return d
}

func (d *ViewData) WithNavItems(items []NavItem) *ViewData {
	d.NavItems = items
	return d
}

func (d *ViewData) WithFlash(level FlashLevel, message string) *ViewData {
	d.Flashes = append(d.Flashes, FlashData{level, message})
	return d
}

func (d *ViewData) WithFlashInfo(message string) *ViewData {
	return d.WithFlash(FlashLevelInfo, message)
}

func (d *ViewData) WithFlashSuccess(message string) *ViewData {
	return d.WithFlash(FlashLevelSuccess, message)
}

func (d *ViewData) WithFlashWarning(message string) *ViewData {
	return d.WithFlash(FlashLevelWarning, message)
}

func (d *ViewData) WithFlashError(message string) *ViewData {
	return d.WithFlash(FlashLevelError, message)
}

func (d *ViewData) WithForm(form *FormData) *ViewData {
	d.Form = form
	return d
}

func (d *ViewData) WithData(data any) *ViewData {
	d.Data = data
	return d
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

func (d *ViewData) Render(code int) {
	d.setFlashCookie(false)

	log := hlog.FromRequest(d.r).With().Str("template", d.templateName).Logger()

	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bytesBufferPool.Put(buf)

	err := d.template.Execute(buf, d)
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
