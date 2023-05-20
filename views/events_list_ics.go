package views

import (
	"net/http"

	ics "github.com/arran4/golang-ical"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog"
	"github.com/xremming/abborre/models"
)

func EventsListICS(cfg aws.Config, tableName string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)

	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		resp, err := models.ListEvents(r.Context(), dynamo, models.ListEventsIn{TableName: tableName})
		if err != nil {
			renderError(w, r, 500, "Failed to list events.")
			return
		}

		cal := ics.NewCalendar()
		for _, event := range resp.Events {
			ev := cal.AddEvent(event.ID())
			ev.SetClass(ics.ClassificationPublic)

			ev.SetCreatedTime(event.Created)
			ev.SetModifiedAt(event.Updated)

			ev.SetDtStampTime(event.StartTime)
			ev.SetStartAt(event.StartTime)
			ev.SetDuration(event.Duration)

			ev.SetSummary(event.Name)
			ev.SetDescription(event.Description)
			// TODO: use base URL
			ev.SetURL("http://localhost:8080/events/" + event.ID())
		}

		w.Header().Set("Content-Type", "text/calendar")
		w.Header().Set("Content-Disposition", "attachment; filename=events.ics")

		err = cal.SerializeTo(w)
		if err != nil {
			log.Err(err).Msg("Failed to serialize calendar")
		}
	}
}
