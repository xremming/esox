package esox

import (
	"context"
	"time"
)

type locationKey struct{}

func GetLocation(ctx context.Context) *time.Location {
	return ctx.Value(locationKey{}).(*time.Location)
}

type nameMappingKey struct{}

func GetNameMapping(ctx context.Context) map[string]URL {
	return ctx.Value(nameMappingKey{}).(map[string]URL)
}
