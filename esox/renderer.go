package esox

import (
	"bytes"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/xremming/abborre/esox/flash"
)

var bytesBufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

type Template struct {
	name     string
	template *template.Template
}

func (t *Template) ViewData(w http.ResponseWriter, r *http.Request) *ViewData {
	flashCookieDeleted := false
	cookie, err := r.Cookie("flash")
	if err == nil {
		flashCookieDeleted = true
		flashes := flash.Decode(cookie.Value)
		r = r.WithContext(flash.NewContext(r.Context(), flashes))
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
