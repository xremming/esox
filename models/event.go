package models

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/xid"
	"github.com/xremming/abborre/esox"
)

type Event struct {
	Base
	Name      string     `dynamodbav:"name"`
	StartTime time.Time  `dynamodbav:"starts"`
	EndTime   *time.Time `dynamodbav:"ends"`
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
		Base:      newBase("event", esox.JoinID("event", id)).withTTL(ttl),
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
