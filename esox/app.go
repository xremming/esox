package esox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
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

type App struct {
	Routes   map[string]http.Handler
	Route404 http.Handler
	CSRF     *csrf.CSRF
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
	c = c.Append(hlog.MethodHandler("method"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("request_id", "X-Request-ID"))
	c = c.Append(hlog.URLHandler("url"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))

	return c
}

func (a *App) Handler(ctx context.Context) http.Handler {
	log := zerolog.Ctx(ctx)

	mux := http.NewServeMux()
	c := a.middleware(*log)

	hasRootPath := false
	for path, handler := range a.Routes {
		if path == "/" && a.Route404 != nil {
			hasRootPath = true
			mux.Handle(path, c.Append(notFoundMiddleware(a.Route404)).Then(handler))
		} else {
			mux.Handle(path, c.Then(handler))
		}
	}

	if !hasRootPath && a.Route404 != nil {
		mux.Handle("/", c.Append(notFoundMiddleware(a.Route404)).Then(http.NotFoundHandler()))
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

func (a *App) Run(ctx context.Context, conf RunConfig) error {
	log := setupLogger(conf.Dev)
	ctx = log.WithContext(ctx)

	if a.CSRF != nil {
		log.Info().Msg("CSRF protection enabled")
		ctx = csrf.NewContext(ctx, a.CSRF)
	} else {
		log.Warn().Msg("CSRF protection disabled")
	}

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
