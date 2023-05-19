package views

import (
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/forms"
	"github.com/xremming/abborre/models"
)

var eventsUpdateTmpl = renderer.GetTemplate("events_update.html", "base.html")

func EventsUpdate(cfg aws.Config, tableName string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)

	updateEventForm := forms.New().
		Field("name", forms.FieldBuilder[forms.TextConfig]{
			Label:  "Name",
			Config: forms.TextConfig{MinLength: 3, MaxLength: 256},
		}).
		Field("startTime", forms.FieldBuilder[forms.DateTimeLocalConfig]{
			Label: "Start Time",
			Config: forms.DateTimeLocalConfig{
				Location: location,
			},
		}).
		Done()

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := xid.FromString(r.URL.Query().Get("id"))
		if err != nil {
			flash.Warning(r, "Invalid event ID.")
			esox.Redirect(w, r, "/events", http.StatusFound)
			return
		}

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				flash.Warning(r, "Something went wrong, please try again.")
				eventsUpdateTmpl.Render(w, r, 400, &data{Nav: defaultNavItems, Form: updateEventForm.Empty()})
				return
			}

			form, parsedForm := updateEventForm.Parse(r.Form)
			if form.HasErrors() {
				eventsUpdateTmpl.Render(w, r, 400, &data{Nav: defaultNavItems, Form: form})
				return
			}

			update := models.UpdateEventIn{TableName: tableName, ID: id}

			name, ok := parsedForm["name"]
			if ok {
				v := name.(string)
				update.Name = &v
			}

			startTime, ok := parsedForm["startTime"]
			if ok {
				v := startTime.(time.Time)
				update.StartTime = &v
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
			flash.Warning(r, "Something went wrong, please try again.")
			eventsUpdateTmpl.Render(w, r, 400, &data{Nav: defaultNavItems, Form: updateEventForm.Empty()})
			return
		}

		form, _ := updateEventForm.Parse(url.Values{
			"name":      {eventOut.Event.Name},
			"startTime": {eventOut.Event.StartTime.Format("2006-01-02T15:04")},
		})

		eventsUpdateTmpl.Render(w, r, 200, &data{Nav: defaultNavItems, Form: form})
	}
}
