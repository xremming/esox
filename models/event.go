package models

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"github.com/xremming/abborre/esox/models"
)

type Event struct {
	models.Base[string, models.ID]
	Name        string        `dynamodbav:"name"`
	Description string        `dynamodbav:"description"`
	StartTime   time.Time     `dynamodbav:"starts,unixtime"`
	Duration    time.Duration `dynamodbav:"duration"`
}

const eventPartitionKey = "event"

var (
	eventName        = expression.Name("name")
	eventDescription = expression.Name("description")
	eventStartTime   = expression.Name("starts")
	eventDuration    = expression.Name("duration")
)

func (e Event) ID() string {
	return e.Base.SortKey.String()
}

type CreateEventIn struct {
	TableName string

	Name        string
	Description string
	StartTime   time.Time
	Duration    time.Duration
}

type CreateEventOut struct {
	Event Event
}

func CreateEvent(ctx context.Context, dynamo *dynamodb.Client, in CreateEventIn) (CreateEventOut, error) {
	id := models.NewID()
	ttl := in.StartTime.Add(180 * 24 * time.Hour)

	event := Event{
		Base:      models.NewBaseWithTTL("event", id, ttl),
		Name:      in.Name,
		StartTime: in.StartTime,
		Duration:  in.Duration,
	}

	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return CreateEventOut{}, err
	}

	zerolog.Ctx(ctx).Debug().
		Interface("event", event).
		Interface("item", item).
		Msg("marshalled item")

	_, err = dynamo.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &in.TableName,
		Item:      item,
	})
	if err != nil {
		return CreateEventOut{}, err
	}

	return CreateEventOut{Event: event}, nil
}

type GetEventIn struct {
	TableName string
	ID        models.ID
}

type GetEventOut struct {
	Event Event
}

func GetEvent(ctx context.Context, dynamo *dynamodb.Client, in GetEventIn) (GetEventOut, error) {
	out, err := dynamo.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &in.TableName,
		Key:       models.GetKey(eventPartitionKey, in.ID),
	})
	if err != nil {
		return GetEventOut{}, err
	}

	event := Event{}
	err = attributevalue.UnmarshalMap(out.Item, &event)
	if err != nil {
		return GetEventOut{}, err
	}

	return GetEventOut{Event: event}, nil
}

type UpdateEventIn struct {
	TableName string

	ID          models.ID
	Name        *string
	Description *string
	StartTime   *time.Time
	Duration    *time.Duration
}

type UpdateEventOut struct {
	Event Event
}

func UpdateEvent(ctx context.Context, dynamo *dynamodb.Client, in UpdateEventIn) (UpdateEventOut, error) {
	log := zerolog.Ctx(ctx).With().Interface("UpdateEventIn", in).Logger()
	log.Info().Msg("UpdateEvent")

	cond := models.NamePartitionKey.Equal(expression.Value(eventPartitionKey))

	update := models.UpdateBuilder(time.Now())

	if in.Name != nil {
		update = update.Set(eventName, expression.Value(in.Name))
	}

	if in.Description != nil {
		update = update.Set(eventDescription, expression.Value(in.Description))
	}

	if in.StartTime != nil {
		update = update.Set(eventStartTime, expression.Value(in.StartTime.Unix()))
	}

	if in.Duration != nil {
		update = update.Set(eventDuration, expression.Value(in.Duration))
	}

	expr, err := expression.NewBuilder().WithCondition(cond).WithUpdate(update).Build()
	if err != nil {
		log.Err(err).Msg("Failed to build expression")
		return UpdateEventOut{}, err
	}

	res, err := dynamo.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &in.TableName,
		Key:                       models.GetKey(eventPartitionKey, in.ID),
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		log.Err(err).Msg("Failed to update event")
		return UpdateEventOut{}, err
	}

	event := Event{}
	err = attributevalue.UnmarshalMap(res.Attributes, &event)
	if err != nil {
		log.Err(err).Msg("Failed to unmarshal event")
		return UpdateEventOut{}, err
	}

	log.Debug().Interface("UpdateEventOut", event).Msg("Updated event")
	return UpdateEventOut{event}, nil
}

type ListEventsIn struct {
	TableName string
}

type ListEventsOut struct {
	Events []Event
}

func ListEvents(ctx context.Context, dynamo *dynamodb.Client, in ListEventsIn) (ListEventsOut, error) {
	keyCond := expression.KeyEqual(
		models.PartitionKey,
		expression.Value(eventPartitionKey),
	)

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return ListEventsOut{}, err
	}

	pager := dynamodb.NewQueryPaginator(dynamo, &dynamodb.QueryInput{
		TableName:                 &in.TableName,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	var items []Event

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return ListEventsOut{}, err
		}
		for _, item := range page.Items {
			out := Event{}
			err = attributevalue.UnmarshalMap(item, &out)
			if err != nil {
				return ListEventsOut{}, err
			}
			items = append(items, out)
		}
	}

	return ListEventsOut{Events: items}, nil
}
