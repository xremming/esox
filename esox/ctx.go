package esox

import (
	"context"
	"io/fs"
	"time"
)

type locationKey struct{}

func GetLocation(ctx context.Context) *time.Location {
	return ctx.Value(locationKey{}).(*time.Location)
}

type staticResourcesKey struct{}

func GetStaticResources(ctx context.Context) fs.FS {
	return ctx.Value(staticResourcesKey{}).(fs.FS)
}

type nameMappingKey struct{}

func GetNameMapping(ctx context.Context) map[string]URL {
	return ctx.Value(nameMappingKey{}).(map[string]URL)
}
