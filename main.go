package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/rs/zerolog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/csrf"
	"github.com/xremming/abborre/views"
)

func GetEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func GetEnvIntOrDefault(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		res, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}

		return res
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
	flagPort      = flag.Int("port", GetEnvIntOrDefault("PORT", 8000), "HTTP port")
	flagTableName = flag.String("table-name", GetEnvOrDefault("TABLE_NAME", "abborre"), "DynamoDB table name")
)

func init() {
	flag.Parse()
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

func main() {
	ctx := context.Background()

	aws, err := setupAWSConfig(ctx, *flagDev)
	if err != nil {
		panic(err)
	}

	app := esox.App{
		StaticResources: os.DirFS("./static/"),
		Routes: map[string]http.Handler{
			"/":                views.Home(),
			"/events":          views.EventsList(aws, *flagTableName),
			"/events/create":   views.EventsCreate(aws, *flagTableName),
			"/events/update":   views.EventsUpdate(aws, *flagTableName),
			"/events/calendar": views.EventsListICS(aws, *flagTableName),
		},
		Handler404: views.NotFound(),
		CSRF: &csrf.CSRF{
			Secrets: []string{"secret"},
		},
	}

	err = app.Run(ctx, esox.RunConfig{
		Dev:  *flagDev,
		Host: *flagHost,
		Port: *flagPort,
	})

	if err != nil {
		panic(err)
	}
}
