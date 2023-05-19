package views

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/models"
)

var eventsListTmpl = renderer2.GetTemplate("events_list.html", "base.html")

func EventsList(cfg aws.Config, tableName string) http.HandlerFunc {
	dynamo := dynamodb.NewFromConfig(cfg)

	return func(w http.ResponseWriter, r *http.Request) {
		log := hlog.FromRequest(r)

		pager := dynamodb.NewQueryPaginator(dynamo, &dynamodb.QueryInput{
			TableName:              &tableName,
			KeyConditionExpression: aws.String("pk = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "event"},
			},
		})

		var items []models.Event

		for pager.HasMorePages() {
			page, err := pager.NextPage(r.Context())
			if err != nil {
				log.Err(err).Msg("Failed to get next page")
				renderError(w, r, 500, "Failed to get next page.")
				return
			}
			for _, item := range page.Items {
				out := models.Event{}
				err = attributevalue.UnmarshalMap(item, &out)
				if err != nil {
					log.Err(err).Interface("item", item).Msg("Failed to unmarshal item")
					renderError(w, r, 500, "Failed to unmarshal item.")
					return
				}
				items = append(items, out)
			}
		}

		eventsListTmpl.Render(w, r, 200, &data{Nav: defaultNavItems, Data: items})
	}
}
