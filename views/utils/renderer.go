package utils

import (
	"bytes"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

var bytesBufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

type Template struct {
	name     string
	template *template.Template
}

func (t *Template) ViewData(w http.ResponseWriter, r *http.Request, title string) *ViewData {
	var flashes []FlashData

	flashCookieDeleted := false
	cookie, err := r.Cookie("flash")
	if err == nil {
		flashCookieDeleted = true
		flashes = decodeFlashCookie(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:    "flash",
			MaxAge:  -1,
			Expires: time.Now().Add(-1 * time.Hour),
		})
	}

	return &ViewData{
		w: w,
		r: r,

		template:     t.template,
		templateName: t.name,

		flashCookieDeleted: flashCookieDeleted,

		Title:   title,
		Flashes: flashes,
	}
}

type Renderer struct {
	FS             fs.FS
	TemplatePrefix string
	BaseTemplate   string
}

func NewRenderer(fs fs.FS, templatePrefix string, baseTemplate string) *Renderer {
	return &Renderer{
		FS:             fs,
		TemplatePrefix: templatePrefix,
		BaseTemplate:   baseTemplate,
	}
}

func (r *Renderer) GetTemplate(name string) *Template {
	return &Template{
		name: name,
		template: template.Must(template.ParseFS(r.FS,
			filepath.Join(r.TemplatePrefix, r.BaseTemplate),
			filepath.Join(r.TemplatePrefix, name),
		)),
	}
}
