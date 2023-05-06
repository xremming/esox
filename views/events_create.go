package views

import (
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/models"
)

type CreateEventForm struct {
	Name      string
	StartTime time.Time
	EndTime   *time.Time
}

func parseCreateEventForm(form url.Values) (CreateEventForm, esox.FormParser) {
	out := CreateEventForm{}
	formParser := esox.NewFormParser()

	location, _ := time.LoadLocation("Europe/Helsinki")

	out.Name = formParser.ParseString(form, "name", esox.ParseStringOpts{
		Required: true,
	})
	out.StartTime = formParser.ParseTime(form, "startTime", esox.ParseTimeOpts{
		Required: true,
		Location: location,
	})
	out.EndTime = formParser.ParseTimePointer(form, "endTime", esox.ParseTimeOpts{
		Location: location,
	})

	return out, *formParser
}

var eventsCreateTmpl = renderer.GetTemplate("events_create.html")

func EventsCreate(cfg aws.Config, tableName *string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)

	return func(w http.ResponseWriter, r *http.Request) {
		log := hlog.FromRequest(r)

		form := &esox.FormData{
			Fields: []esox.FormField{
				{Name: "name"},
				{Name: "startTime"},
				{Name: "endTime"},
			},
		}

		d := eventsCreateTmpl.ViewData(w, r, "EventsCreate").
			WithNavItems(defaultNavItems).
			WithForm(form)

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				renderError(w, r, 400, "Failed to parse form.")
				return
			}

			parsedForm, formParser := parseCreateEventForm(r.Form)
			if formParser.HasErrors() {
				formParser.UpdateForm(form)
				log.Info().
					Interface("form", form).
					Interface("formParser", formParser).
					Msg("formParser has errors")

				d.Render(400)
				return
			}

			log.Info().Interface("form", form).Msg("Create event")

			_, err = models.CreateEvent(r.Context(), dynamo, models.CreateEventIn{
				TableName: *tableName,
				Name:      parsedForm.Name,
				StartTime: parsedForm.StartTime,
				EndTime:   parsedForm.EndTime,
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
