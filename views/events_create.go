package views

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/models"
)

var eventsCreateTmpl = renderer.GetTemplate("events_create.html")

func EventsCreate(cfg aws.Config, tableName *string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)
	// tmpl := getTemplate("events_create.html")
	_ = dynamo

	return func(w http.ResponseWriter, r *http.Request) {
		log := hlog.FromRequest(r)

		d := eventsCreateTmpl.ViewData(w, r, "EventsCreate").
			WithNavItems(defaultNavItems)

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				renderError(w, r, 400, "Failed to parse form.")
				return
			}

			name := r.Form.Get("name")

			startTime := r.Form.Get("startTime")
			startTimeParsed, err := time.Parse("2006-01-02T15:04", startTime)
			if err != nil {
				log.Err(err).Str("startTime", startTime).Msg("Failed to parse start time")
				renderError(w, r, 400, "Failed to parse start time.")
				return
			}

			log.Info().Str("name", name).Str("startTime", startTime).Msg("Create event")

			_, err = models.CreateEvent(r.Context(), dynamo, models.CreateEventIn{
				TableName: *tableName,
				Name:      name,
				StartTime: startTimeParsed,
			})
			if err != nil {
				log.Err(err).Msg("Failed to create event")
				renderError(w, r, 500, "Failed to create event.")
				return
			}

			d.WithFlashSuccess("Event created.").
				Redirect("/events", http.StatusFound)
		}

		d.Render(200)
	}
}
