package models

import "time"

type Base struct {
	PartitionKey string `dynamodbav:"pk"`
	SortKey      string `dynamodbav:"sk"`
	Created      int64  `dynamodbav:"created"`
	Updated      int64  `dynamodbav:"updated"`
	TimeToLive   *int64 `dynamodbav:"ttl"`
}

func newBase(pk, sk string) Base {
	now := time.Now().Unix()
	return Base{
		PartitionKey: pk,
		SortKey:      sk,
		Created:      now,
		Updated:      now,
	}
}

func (b Base) withTTL(ttl time.Time) Base {
	ttlUnix := ttl.Unix()
	b.TimeToLive = &ttlUnix
	return b
}
