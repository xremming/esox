package views

import (
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/forms"
	esoxModels "github.com/xremming/abborre/esox/models"
	"github.com/xremming/abborre/models"
)

var eventsUpdateTmpl = esox.GetTemplate("events_update.html", "base.html")

func EventsUpdate(cfg aws.Config, tableName string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)

	updateEventForm := forms.New().
		Field("name", forms.FieldBuilder[forms.TextConfig]{
			Label:  "Name",
			Config: forms.TextConfig{MinLength: 3, MaxLength: 256},
		}).
		Field("description", forms.FieldBuilder[forms.TextConfig]{
			Label:  "Description",
			Config: forms.TextConfig{Multiline: true, MinLength: 3},
		}).
		Field("startTime", forms.FieldBuilder[forms.DateTimeLocalConfig]{
			Label: "Start Time",
		}).
		Field("duration", forms.FieldBuilder[forms.TextConfig]{
			Label:  "Duration",
			Config: forms.TextConfig{Parse: forms.ParseDuration},
		}).
		Done()

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := esoxModels.IDFromString(r.URL.Query().Get("id"))
		if err != nil {
			flash.Warning(r, "Invalid event ID.")
			esox.Redirect(w, r, "/events", http.StatusFound)
			return
		}

		log := zerolog.Ctx(r.Context()).With().Str("id", id.String()).Logger()

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				flash.Warning(r, "Something went wrong, please try again.")
				eventsUpdateTmpl.Render(w, r, 400, &data{Form: updateEventForm.Empty(r.Context())})
				return
			}

			form, parsedForm := updateEventForm.Parse(r.Context(), r.Form)
			if form.HasErrors() {
				eventsUpdateTmpl.Render(w, r, 400, &data{Form: form})
				return
			}

			update := models.UpdateEventIn{TableName: tableName, ID: id}

			name, ok := parsedForm["name"]
			if ok {
				v := name.(string)
				update.Name = &v
			}

			description, ok := parsedForm["description"]
			if ok {
				v := description.(string)
				update.Description = &v
			}

			startTime, ok := parsedForm["startTime"]
			if ok {
				v := startTime.(time.Time)
				update.StartTime = &v
			}

			duration, ok := parsedForm["duration"]
			if ok {
				v := duration.(time.Duration)
				update.Duration = &v
			}

			_, err = models.UpdateEvent(r.Context(), dynamo, update)
			if err != nil {
				log.Error().Err(err).Msg("Failed to update event")
				flash.Warning(r, "Something went wrong, please try again.")
				esox.Redirect(w, r, "/events", http.StatusFound)
				return
			}

			esox.Redirect(w, r, "/events", http.StatusFound)
			return
		}

		eventOut, err := models.GetEvent(r.Context(), dynamo, models.GetEventIn{TableName: tableName, ID: id})
		if err != nil {
			log.Err(err).Msg("Failed to get event")

			flash.Warning(r, "Something went wrong, please try again.")
			eventsUpdateTmpl.Render(w, r, 400, &data{
				Form: updateEventForm.Empty(r.Context()),
			})
			return
		}

		location, err := time.LoadLocation("Europe/Helsinki")
		if err != nil {
			log.Err(err).Str("location", "Europe/Helsinki").Msg("Failed to load location")
			renderError(w, r, 500, "Something went wrong, please try again.")
			return
		}

		form := updateEventForm.Prefilled(r.Context(), url.Values{
			"name":        {eventOut.Event.Name},
			"description": {eventOut.Event.Description},
			"startTime":   {eventOut.Event.StartTime.In(location).Format(forms.FormatDatetimeLocal)},
			"duration":    {eventOut.Event.Duration.String()},
		})

		eventsUpdateTmpl.Render(w, r, 200, &data{Form: form})
	}
}
