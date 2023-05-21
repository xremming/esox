package views

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/forms"
	"github.com/xremming/abborre/models"
)

var eventsCreateTmpl = esox.GetTemplate("events_create.html", "base.html")

func EventsCreate(cfg aws.Config, tableName string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)

	createEventForm := forms.New().
		Field("name", forms.FieldBuilder[forms.TextConfig]{
			Label:    "Name",
			Required: true,
			Config:   forms.TextConfig{MinLength: 3, MaxLength: 256},
		}).
		Field("description", forms.FieldBuilder[forms.TextConfig]{
			Label:  "Description",
			Config: forms.TextConfig{Multiline: true, MinLength: 3},
		}).
		Field("startTime", forms.FieldBuilder[forms.DateTimeLocalConfig]{
			Label:    "Start Time",
			Required: true,
		}).
		Field("duration", forms.FieldBuilder[forms.SelectConfig]{
			Label:    "Duration",
			Required: true,
			Config: forms.SelectConfig{
				Parse: forms.ParseDuration,
				Options: []forms.OptionConfig{
					{Value: "15m", Label: "15 minutes"},
					{Value: "30m", Label: "30 minutes"},
					{Value: "45m", Label: "45 minutes"},
					{Value: "1h", Label: "1 hour", Selected: true},
					{Value: "2h", Label: "2 hours"},
					{Value: "3h", Label: "3 hours"},
					{Value: "4h", Label: "4 hours"},
					{Value: "5h", Label: "5 hours"},
					{Value: "6h", Label: "6 hours"},
				},
			},
		}).
		Done()

	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				flash.Warning(r, "Something went wrong, please try again.")
				eventsCreateTmpl.Render(w, r, 400, &data{Form: createEventForm.Empty(r.Context())})
				return
			}

			form, parsedForm := createEventForm.Parse(r.Context(), r.Form)
			if form.HasErrors() {
				log.Info().
					Interface("form", form).
					Msg("parsed form has errors")

				flash.Error(r, "Failed to create new event.")
				eventsCreateTmpl.Render(w, r, 400, &data{Form: form})
				return
			}

			log.Info().Interface("data", parsedForm).Msg("Create event")

			name := parsedForm["name"].(string)
			description := parsedForm["description"].(string)
			startTime := parsedForm["startTime"].(time.Time)
			duration := parsedForm["duration"].(time.Duration)

			createEvent := models.CreateEventIn{
				TableName:   tableName,
				Name:        name,
				Description: description,
				StartTime:   startTime,
				Duration:    duration,
			}

			_, err = models.CreateEvent(r.Context(), dynamo, createEvent)
			if err != nil {
				log.Err(err).Msg("Failed to create event")
				flash.Error(r, "Failed to create event.")
				eventsCreateTmpl.Render(w, r, 500, &data{Form: form})
				return
			}

			flash.Info(r, "Event created.")
			esox.Redirect(w, r, "/events", http.StatusFound)
		} else {
			eventsCreateTmpl.Render(w, r, 200, &data{Form: createEventForm.Empty(r.Context())})
		}
	}
}
