package csrf

import (
	"context"
)

type csrfKey struct{}

func NewContext(ctx context.Context, csrf *CSRF) context.Context {
	return context.WithValue(ctx, csrfKey{}, csrf)
}

func FromContext(ctx context.Context) *CSRF {
	value := ctx.Value(csrfKey{})
	if value == nil {
		return nil
	}

	return value.(*CSRF)
}
