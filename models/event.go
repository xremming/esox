package models

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"github.com/xremming/abborre/esox"
)

type Event struct {
	Base
	Name      string     `dynamodbav:"name"`
	StartTime time.Time  `dynamodbav:"starts,unixtime"`
	EndTime   *time.Time `dynamodbav:"ends,unixtime,omitempty"`
}

func (e Event) ID() xid.ID {
	splitted := strings.SplitN(e.Base.SortKey, ":", 2)
	if len(splitted) != 2 {
		return xid.ID{}
	}

	id, err := xid.FromString(splitted[1])
	if err != nil {
		return xid.ID{}
	}

	return id
}

type CreateEventIn struct {
	TableName string

	Name      string
	StartTime time.Time
	EndTime   *time.Time
}

type CreateEventOut struct {
	Event Event
}

func CreateEvent(ctx context.Context, dynamo *dynamodb.Client, in CreateEventIn) (CreateEventOut, error) {
	id := xid.New()
	ttl := in.StartTime.Add(180 * 24 * time.Hour)

	event := Event{
		Base:      newBaseWithTTL("event", esox.JoinID("event", id), ttl),
		Name:      in.Name,
		StartTime: in.StartTime,
		EndTime:   in.EndTime,
	}

	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return CreateEventOut{}, err
	}

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
	ID        xid.ID
}

type GetEventOut struct {
	Event Event
}

func GetEvent(ctx context.Context, dynamo *dynamodb.Client, in GetEventIn) (GetEventOut, error) {
	out, err := dynamo.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &in.TableName,
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "event"},
			"sk": &types.AttributeValueMemberS{Value: esox.JoinID("event", in.ID)},
		},
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

	ID        xid.ID
	Name      *string
	StartTime *time.Time
	EndTime   *time.Time
}

type UpdateEventOut struct {
	Event Event
}

func UpdateEvent(ctx context.Context, dynamo *dynamodb.Client, in UpdateEventIn) (UpdateEventOut, error) {
	log := log.Ctx(ctx).With().Interface("UpdateEventIn", in).Logger()
	log.Info().Msg("UpdateEvent")

	cond := expression.Name("pk").Equal(expression.Value("event"))

	update := baseUpdate(time.Now())

	if in.Name != nil {
		update = update.Set(expression.Name("name"), expression.Value(in.Name))
	}

	if in.StartTime != nil {
		update = update.Set(expression.Name("starts"), expression.Value(in.StartTime.Unix()))
	}

	if in.EndTime != nil {
		update = update.Set(expression.Name("ends"), expression.Value(in.EndTime.Unix()))
	}

	expr, err := expression.NewBuilder().WithCondition(cond).WithUpdate(update).Build()
	if err != nil {
		log.Err(err).Msg("Failed to build expression")
		return UpdateEventOut{}, err
	}

	res, err := dynamo.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &in.TableName,
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "event"},
			"sk": &types.AttributeValueMemberS{Value: esox.JoinID("event", in.ID)},
		},
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
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
	keyCond := expression.KeyEqual(expression.Key("pk"), expression.Value("event")).
		And(expression.KeyBeginsWith(expression.Key("sk"), "event"))

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
