package esox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/utils"
)

type Template struct {
	name          string
	baseName      string
	baseTemplate  string
	childTemplate string
}

const TemplatesPrefix = "templates"

func (t *Template) reload() {
	baseFile, err := os.Open(filepath.Join(TemplatesPrefix, t.baseName))
	if err != nil {
		panic(err)
	}
	defer baseFile.Close()

	childFile, err := os.Open(filepath.Join(TemplatesPrefix, t.name))
	if err != nil {
		panic(err)
	}
	defer childFile.Close()

	baseContent, err := io.ReadAll(baseFile)
	if err != nil {
		panic(err)
	}
	t.baseTemplate = string(baseContent)

	childContent, err := io.ReadAll(childFile)
	if err != nil {
		panic(err)
	}
	t.childTemplate = string(childContent)
}

func GetTemplate(name, baseName string) *Template {
	out := Template{name: name, baseName: baseName}
	out.reload()

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
	ModTime() (lastModified time.Time)
	CacheControl() (public bool, maxAge time.Duration)
	SetFlashes(flashes []flash.Data)
}

func (t *Template) funcs(ctx context.Context) template.FuncMap {
	return template.FuncMap{
		"now": func() time.Time {
			location := GetLocation(ctx)
			return time.Now().In(location)
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
		"partial": func(name string, data interface{}) (template.HTML, error) {
			file, err := os.Open(filepath.Join(TemplatesPrefix, name))
			if err != nil {
				return "", err
			}
			defer file.Close()

			buf := utils.GetBytesBuffer()
			defer utils.PutBytesBuffer(buf)

			_, err = io.Copy(buf, file)
			if err != nil {
				return "", err
			}

			tmpl, err := template.New(name).Funcs(t.funcs(ctx)).Parse(buf.String())
			if err != nil {
				return "", err
			}

			// reuse the buffer for the output
			buf.Reset()

			err = tmpl.Execute(buf, data)
			if err != nil {
				return "", err
			}

			return template.HTML(buf.String()), nil
		},
		"stylesheet": func(name string) (template.HTML, error) {
			file, err := GetStaticFile(name)
			if err != nil {
				return "", err
			}
			defer file.Close()

			return template.HTML(fmt.Sprintf(
				`<link rel="stylesheet" href="/static/%s" integrity="%s">`,
				template.HTMLEscapeString(file.PathWithHash), file.Integrity,
			)), nil
		},
		"javascript": func(name string) (template.HTML, error) {
			file, err := GetStaticFile(name)
			if err != nil {
				return "", err
			}
			defer file.Close()

			return template.HTML(fmt.Sprintf(
				`<script src="/static/%s" integrity="%s" async></script>`,
				template.HTMLEscapeString(file.PathWithHash), file.Integrity,
			)), nil
		},
		"urlFor": func(name string) (string, error) {
			nameMapping := GetNameMapping(ctx)
			url, ok := nameMapping[name]
			if !ok {
				return "", fmt.Errorf("unknown route name: %s", name)
			}

			return url.Path, nil
		},
		"urlForStatic": func(name string) (string, error) {
			file, err := GetStaticFile(name)
			if err != nil {
				return "", err
			}
			defer file.Close()

			return fmt.Sprintf("/static/%s", file.PathWithHash), nil
		},
	}
}

func (t *Template) Render(w http.ResponseWriter, r *http.Request, code int, data RenderData) {
	runConfig := GetRunConfig(r.Context())
	if runConfig.Dev {
		t.reload()
	}

	log := hlog.FromRequest(r).With().
		Int("code", code).
		Str("template", t.name).
		Interface("data", data).
		Logger()

	flashes := flash.FromRequest(r)
	setFlashCookie(w, r, false, flashes)
	data.SetFlashes(flashes)

	tmpl, err := template.New(t.baseName).
		Funcs(t.funcs(r.Context())).
		Parse(t.baseTemplate)
	if err != nil {
		log.Err(err).Msg("failed to parse base template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err = tmpl.Parse(t.childTemplate)
	if err != nil {
		log.Err(err).Msg("failed to parse child template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(buf)

	h := sha256.New()
	out := io.MultiWriter(buf, h)

	err = tmpl.Execute(out, data)
	if err != nil {
		log.Err(err).Msg("failed to execute template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// When the code is not 200 we cannot be sure whether the content may be cached.
	if code == http.StatusOK {
		etag := fmt.Sprintf(`"%s"`, base64.URLEncoding.EncodeToString(h.Sum(nil)))
		w.Header().Set("ETag", etag)

		public, maxAge := data.CacheControl()
		if maxAge > 0 {
			prefix := "private "
			if public {
				prefix = ""
			}

			cacheControl := fmt.Sprintf("%smax-age=%d", prefix, int(maxAge.Seconds()))
			w.Header().Set("Cache-Control", cacheControl)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// The http.ServeContent function is only guaranteed to work correctly when the status code is 200.
	if code == http.StatusOK {
		http.ServeContent(w, r, t.name, data.ModTime(), bytes.NewReader(buf.Bytes()))
		return
	}

	w.WriteHeader(code)
	_, err = io.Copy(w, buf)
	if err != nil {
		log.Err(err).Msg("Failed to write response body.")
	}
}

func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	setFlashCookie(w, r, true, flash.FromRequest(r))
	http.Redirect(w, r, url, code)
}
