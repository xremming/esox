package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/rs/zerolog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/csrf"
	"github.com/xremming/abborre/views"
	"go-simpler.org/env"
	"golang.org/x/oauth2"
)

type Config struct {
	Dev               bool     `env:"DEV"`
	Host              string   `env:"HOST" default:"localhost"`
	Port              int      `env:"PORT" default:"8000"`
	TableName         string   `env:"TABLE_NAME,required"`
	AdminPassword     string   `env:"ADMIN_PASSWORD,required"`
	BaseURL           string   `env:"BASE_URL,required"`
	Secrets           []string `env:"SECRETS,required"`
	OAuthClientID     string   `env:"OAUTH_CLIENT_ID,required"`
	OAuthClientSecret string   `env:"OAUTH_CLIENT_SECRET,required"`
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

	var cfg Config
	err := env.Load(&cfg)
	if err != nil {
		panic(err)
	}

	aws, err := setupAWSConfig(ctx, cfg.Dev)
	if err != nil {
		panic(err)
	}

	oAuth2Config := oauth2.Config{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: cfg.OAuthClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://discord.com/oauth2/authorize",
			TokenURL:  "https://discord.com/api/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: cfg.BaseURL + "/discord/callback",
		Scopes:      []string{"identify", "email", "guilds.join"},
	}

	auth := esox.BasicAuth("admin", cfg.AdminPassword)

	app := esox.App{
		BaseURL:         cfg.BaseURL,
		StaticResources: os.DirFS("./static/"),
		Routes: map[string]http.Handler{
			"/":                    views.Home(),
			"/events":              views.EventsList(aws, cfg.TableName),
			"/events/calendar":     views.EventsListICS(aws, cfg.TableName),
			"/admin/events/create": auth(views.EventsCreate(aws, cfg.TableName)),
			"/admin/events/update": auth(views.EventsUpdate(aws, cfg.TableName)),
			"/discord/login":       views.DiscordLogin(oAuth2Config),
			"/discord/callback":    views.DiscordCallback(oAuth2Config),
		},
		Handler404: views.NotFound(),
		CSRF: &csrf.CSRF{
			Secrets: cfg.Secrets,
		},
	}

	err = app.Run(ctx, esox.RunConfig{
		Dev:  cfg.Dev,
		Host: cfg.Host,
		Port: cfg.Port,
	})

	if err != nil {
		panic(err)
	}
}
