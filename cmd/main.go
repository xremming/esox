package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/views"
)

func GetEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func GetEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		res, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}

		return res
	}

	return defaultValue
}

var (
	flagDev       = flag.Bool("dev", GetEnvBoolOrDefault("DEV", false), "Development mode")
	flagHost      = flag.String("host", GetEnvOrDefault("HOST", "localhost"), "HTTP host")
	flagPort      = flag.String("port", GetEnvOrDefault("PORT", "8000"), "HTTP port")
	flagTableName = flag.String("table-name", GetEnvOrDefault("TABLE_NAME", "abborre"), "DynamoDB table name")
)

func init() {
	flag.Parse()
}

func setupAWSConfig(ctx context.Context) (aws.Config, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log := zerolog.Ctx(ctx)

	if *flagDev {
		log.Info().Msg("Using local DynamoDB")

		return config.LoadDefaultConfig(ctx,
			config.WithRegion("eu-west-1"),
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

func main() {
	// Setup logger.
	var w io.Writer = os.Stderr
	if *flagDev {
		w = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	log := zerolog.New(w).With().Timestamp().Caller().Logger()
	ctx := log.WithContext(context.Background())

	// Setup default middleware.
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

	// Setup AWS config.
	cfg, err := setupAWSConfig(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to configure AWS")
	}

	// Setup routes and HTTP server.
	mux := http.ServeMux{}
	mux.Handle("/static/", c.Then(views.Static()))
	mux.Handle("/events/create", c.Then(views.EventsCreate(cfg, flagTableName)))
	mux.Handle("/events", c.Then(views.EventsList(cfg, flagTableName)))
	mux.Handle("/", c.ThenFunc(views.Home()))

	addr := *flagHost + ":" + *flagPort
	srv := &http.Server{
		Addr:    addr,
		Handler: &mux,
	}

	// Start server in goprocess.
	go func() {
		log.Info().
			Bool("dev", *flagDev).
			Str("tableName", *flagTableName).
			Str("addr", addr).
			Msg("HTTP server starting")

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("HTTP server closed")
		} else {
			log.Err(err).Msg("HTTP server ListenAndServe")
		}
	}()

	// Wait for a signal to quit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Shutdown the server.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
