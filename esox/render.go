package esox

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/utils"
)

type Renderer struct {
	fs     fs.FS
	prefix string
}

func NewRenderer(fs fs.FS, prefix string) *Renderer {
	return &Renderer{fs, prefix}
}

type Page struct {
	baseTemplate  string
	baseName      string
	name          string
	childTemplate string
}

func (re *Renderer) GetTemplate(name, baseName string) *Page {
	out := Page{name: name, baseName: baseName}

	baseFile, err := re.fs.Open(filepath.Join(re.prefix, baseName))
	if err != nil {
		panic(err)
	}
	defer baseFile.Close()

	childFile, err := re.fs.Open(filepath.Join(re.prefix, name))
	if err != nil {
		panic(err)
	}
	defer childFile.Close()

	baseContent, err := io.ReadAll(baseFile)
	if err != nil {
		panic(err)
	}
	out.baseTemplate = string(baseContent)

	childContent, err := io.ReadAll(childFile)
	if err != nil {
		panic(err)
	}
	out.childTemplate = string(childContent)

	return &out
}

func checkFlashCookie(w http.ResponseWriter, r *http.Request) bool {
	flashCookieDeleted := false
	cookie, err := r.Cookie("flash")
	if err == nil {
		flashCookieDeleted = true
		flashes := flash.Decode(cookie.Value)
		*r = *r.WithContext(flash.NewContext(r.Context(), flashes))
		http.SetCookie(w, &http.Cookie{
			Name:    "flash",
			MaxAge:  -1,
			Expires: time.Now().Add(-1 * time.Hour),
		})
	}

	return flashCookieDeleted
}

func setFlashCookie(w http.ResponseWriter, r *http.Request, redirect bool, flashes []flash.Data) {
	flashCookieDeleted := checkFlashCookie(w, r)
	if !redirect || flashCookieDeleted {
		return
	}

	if len(flashes) == 0 {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    flash.Encode(flashes),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

type RenderData interface {
	SetFlashes(flashes []flash.Data)
}

func sha256sum(file io.Reader) (string, error) {
	hash := sha256.New()

	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	sum := hash.Sum(nil)
	return fmt.Sprintf("sha256-%s", base64.StdEncoding.EncodeToString(sum)), nil
}

func (p *Page) Funcs(ctx context.Context) template.FuncMap {
	staticResources := GetStaticResources(ctx)
	location := GetLocation(ctx)

	now := time.Now().In(location)
	return template.FuncMap{
		"now": func() time.Time {
			return now
		},
		"time": func(t time.Time, value string) template.HTML {
			return template.HTML(fmt.Sprintf(
				`<time datetime="%s">%s</time>`,
				t.Format(time.RFC3339),
				template.HTMLEscapeString(value),
			))
		},
		"formatTime": func(layout string, t time.Time) string {
			return t.Format(layout)
		},
		"stylesheet": func(name string) (template.HTML, error) {
			file, err := staticResources.Open(name)
			if err != nil {
				return "", err
			}

			hash, err := sha256sum(file)
			if err != nil {
				return "", err
			}

			return template.HTML(fmt.Sprintf(
				`<link rel="stylesheet" href="/static/%s" integrity="%s">`,
				name, hash,
			)), nil
		},
		"javascript": func(name string) (template.HTML, error) {
			file, err := staticResources.Open(name)
			if err != nil {
				return "", err
			}

			hash, err := sha256sum(file)
			if err != nil {
				return "", err
			}

			return template.HTML(fmt.Sprintf(
				`<script async src="/static/%s" integrity="%s"></script>`,
				name, hash,
			)), nil
		},
	}
}

func (p *Page) Render(w http.ResponseWriter, r *http.Request, code int, data RenderData) {
	log := hlog.FromRequest(r).With().
		Int("code", code).
		Str("template", p.name).
		Interface("data", data).
		Logger()

	flashes := flash.FromRequest(r)
	setFlashCookie(w, r, false, flashes)
	data.SetFlashes(flashes)

	tmpl, err := template.New(p.baseName).
		Funcs(p.Funcs(r.Context())).
		Parse(p.baseTemplate)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse base template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err = tmpl.Parse(p.childTemplate)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse child template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(buf)

	err = tmpl.Execute(buf, data)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Error().Err(err).Msg("failed to write template to response")
	}
}

func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	setFlashCookie(w, r, true, flash.FromRequest(r))
	http.Redirect(w, r, url, code)
}
