package flash

import (
	"context"
	"net/http"
)

type flashKey struct{}

func NewContext(ctx context.Context, flashes []Data) context.Context {
	return context.WithValue(ctx, flashKey{}, flashes)
}

func FromContext(ctx context.Context) []Data {
	value := ctx.Value(flashKey{})
	if value == nil {
		return nil
	}

	return value.([]Data)
}

func FromRequest(r *http.Request) []Data {
	return FromContext(r.Context())
}
