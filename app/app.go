package app

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/views"
)

func setupMiddleware(log zerolog.Logger) alice.Chain {
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
	c = c.Append(hlog.URLHandler("url"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			next.ServeHTTP(w, r)
		})
	})

	return c
}

func setupAWSConfig(ctx context.Context, isDev bool) (aws.Config, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log := zerolog.Ctx(ctx)

	if isDev {
		log.Info().Msg("Using local DynamoDB")

		return config.LoadDefaultConfig(ctx,
			config.WithRegion("eu-north-1"),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, opts ...any) (aws.Endpoint, error) {
					return aws.Endpoint{URL: "http://localhost:8000"}, nil
				}),
			),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     "local",
					SecretAccessKey: "verysecret",
				},
			}),
		)
	}

	return config.LoadDefaultConfig(ctx)
}

type Configuration struct {
	IsDev     bool
	TableName string
}

func NewHandler(ctx context.Context, conf Configuration) *http.ServeMux {
	log := zerolog.Ctx(ctx)

	cfg, err := setupAWSConfig(ctx, conf.IsDev)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to configure AWS")
	}

	c := setupMiddleware(*log)
	mux := http.ServeMux{}
	mux.Handle("/static/", c.Then(views.Static()))
	mux.Handle("/events/create", c.Then(views.EventsCreate(cfg, conf.TableName)))
	mux.Handle("/events", c.Then(views.EventsList(cfg, conf.TableName)))
	mux.Handle("/", c.ThenFunc(views.Home()))

	return &mux
}
