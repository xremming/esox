package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type KeyType interface {
	~string | ID | NSID
}

type Base[PK, SK KeyType] struct {
	PartitionKey PK         `dynamodbav:"pk"`
	SortKey      SK         `dynamodbav:"sk"`
	Created      time.Time  `dynamodbav:"created,unixtime"`
	Updated      time.Time  `dynamodbav:"updated,unixtime"`
	Version      int        `dynamodbav:"version"`
	TimeToLive   *time.Time `dynamodbav:"ttl,unixtime,omitempty"`
}

func (b Base[PK, SK]) Key() map[string]types.AttributeValue {
	return Key(b.PartitionKey, b.SortKey)
}

var (
	PartitionKey = expression.Key("pk")
	SortKey      = expression.Key("sk")
)

var (
	NamePartitionKey = expression.Name("pk")
	NameSortKey      = expression.Name("sk")
	NameCreated      = expression.Name("created")
	NameUpdated      = expression.Name("updated")
	NameVersion      = expression.Name("version")
	NameTimeToLive   = expression.Name("ttl")
)

func Key[PK, SK KeyType](pk PK, sk SK) map[string]types.AttributeValue {
	out, err := attributevalue.MarshalMap(map[string]any{
		"pk": &pk,
		"sk": &sk,
	})
	if err != nil {
		panic(err)
	}

	return out
}

func NewBase[PK, SK KeyType](pk PK, sk SK) Base[PK, SK] {
	now := time.Now().UTC()
	return Base[PK, SK]{
		PartitionKey: pk,
		SortKey:      sk,
		Created:      now,
		Updated:      now,
		Version:      1,
	}
}

func NewBaseWithTTL[PK, SK KeyType](pk PK, sk SK, ttl time.Time) Base[PK, SK] {
	out := NewBase(pk, sk)
	out.TimeToLive = &ttl
	return out
}

func UpdateBuilder(now time.Time) expression.UpdateBuilder {
	update := expression.UpdateBuilder{}

	update = update.Set(NameCreated,
		expression.IfNotExists(NameCreated, expression.Value(now.UTC().Unix())),
	)

	update = update.Set(NameUpdated,
		expression.Value(now.UTC().Unix()),
	)

	update = update.Set(NameVersion,
		expression.Plus(
			expression.IfNotExists(NameVersion, expression.Value(0)),
			expression.Value(1),
		),
	)

	return update
}

func UpdateTTL(update expression.UpdateBuilder, ttl *time.Time) expression.UpdateBuilder {
	if ttl != nil {
		return update.Set(NameTimeToLive,
			expression.Value(ttl.UTC().Unix()),
		)
	} else {
		return update.Remove(NameTimeToLive)
	}
}
