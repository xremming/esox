package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

type Base struct {
	PartitionKey string     `dynamodbav:"pk"`
	SortKey      string     `dynamodbav:"sk"`
	Created      time.Time  `dynamodbav:"created,unixtime"`
	Updated      time.Time  `dynamodbav:"updated,unixtime"`
	TimeToLive   *time.Time `dynamodbav:"ttl,unixtime,omitempty"`
}

func newBase(pk, sk string) Base {
	now := time.Now().UTC()
	return Base{
		PartitionKey: pk,
		SortKey:      sk,
		Created:      now,
		Updated:      now,
	}
}

func newBaseWithTTL(pk, sk string, ttl time.Time) Base {
	out := newBase(pk, sk)
	out.TimeToLive = &ttl
	return out
}

func baseUpdate(now time.Time) expression.UpdateBuilder {
	return expression.Set(
		expression.Name("updated"),
		expression.Value(now.UTC().Unix()),
	)
}
