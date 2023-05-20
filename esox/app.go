package esox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox/csrf"
)

func notFoundMiddleware(notFound http.Handler) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				notFound.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type XFrameOptions string

const (
	XFrameOptionsDeny       XFrameOptions = "DENY"
	XFrameOptionsSameOrigin XFrameOptions = "SAMEORIGIN"
)

type Security struct {
	XFrameOptions XFrameOptions
	NoSniff       bool
	CSP           string
	// TODO: HSTS
}

var DefaultSecurity = Security{
	XFrameOptions: XFrameOptionsDeny,
	NoSniff:       true,
	CSP:           "default-src 'self'",
}

type App struct {
	StaticResources fs.FS
	Routes          map[string]http.Handler
	Handler404      http.Handler
	CSRF            *csrf.CSRF
	Security        *Security
}

func (a *App) middleware(log zerolog.Logger) alice.Chain {
	c := alice.New()

	c = c.Append(hlog.NewHandler(log))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(
		hlog.MethodHandler("method"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("request_id", "X-Request-ID"),
		hlog.URLHandler("url"),
		hlog.UserAgentHandler("user_agent"),
	)
	c = c.Append(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			security := DefaultSecurity
			if a.Security != nil {
				security = *a.Security
			}

			if security.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", string(security.XFrameOptions))
			}

			if security.NoSniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			if security.CSP != "" {
				w.Header().Set("Content-Security-Policy", security.CSP)
			}

			next.ServeHTTP(w, r)
		})
	})

	return c
}

func (a *App) Handler(ctx context.Context) http.Handler {
	log := zerolog.Ctx(ctx)

	mux := http.NewServeMux()
	c := a.middleware(*log)

	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/static/")
		logStatic := log.With().Str("path", r.URL.Path).Logger()

		file, err := a.StaticResources.Open(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				http.NotFound(w, r)
			} else {
				logStatic.Err(err).Msg("error opening static resource")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}

			return
		}
		defer file.Close()

		var buf [512]byte
		n, err := file.Read(buf[:])
		if err != nil && err != io.EOF {
			logStatic.Err(err).Msg("error reading static resource")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		contentType := ""
		if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		} else {
			contentType = http.DetectContentType(buf[:n])
		}

		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		_, err = w.Write(buf[:n])
		if err != nil {
			logStatic.Err(err).Msg("failed to write head of static resource to response writer")
			return
		}

		_, err = io.Copy(w, file)
		if err != nil {
			logStatic.Err(err).Msg("failed to copy static resource to response writer")
			return
		}
	}

	mux.Handle("/static/", c.ThenFunc(staticHandler))

	hasRootPath := false
	for path, handler := range a.Routes {
		if path == "/static/" {
			panic("reserved path: /static/")
		}

		if path == "/" && a.Handler404 != nil {
			hasRootPath = true
			mux.Handle(path, c.Append(notFoundMiddleware(a.Handler404)).Then(handler))
		} else {
			mux.Handle(path, c.Then(handler))
		}
	}

	if !hasRootPath && a.Handler404 != nil {
		mux.Handle("/", c.Append(notFoundMiddleware(a.Handler404)).Then(http.NotFoundHandler()))
	}

	return mux
}

const (
	DefaultShutdownTimeout = 5 * time.Second
)

type RunConfig struct {
	Dev             bool
	Host            string
	Port            int
	ShutdownTimeout time.Duration
}

type staticResourcesKey struct{}

func (a *App) Run(ctx context.Context, conf RunConfig) error {
	log := setupLogger(conf.Dev)
	ctx = log.WithContext(ctx)

	if a.CSRF != nil {
		log.Info().Msg("CSRF protection enabled")
		ctx = csrf.NewContext(ctx, a.CSRF)
	} else {
		log.Warn().Msg("CSRF protection disabled")
	}

	ctx = context.WithValue(ctx, staticResourcesKey{}, a.StaticResources)

	handler := a.Handler(ctx)

	// If AWS_LAMBDA_RUNTIME_API is set, start the Lambda runtime API instead.
	if _, ok := os.LookupEnv("AWS_LAMBDA_RUNTIME_API"); ok {
		lambdaurl.Start(handler)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		log.Info().
			Str("addr", addr).
			Msg("HTTP server starting")

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("HTTP server closed")
		} else {
			log.Err(err).Msg("HTTP server ListenAndServe failed")
		}
	}()

	// Wait for a signal to quit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Shutdown the server.
	t := conf.ShutdownTimeout
	if t == 0 {
		t = DefaultShutdownTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Err(err).Msg("HTTP server shutdown had an error")
	}

	return nil
}
