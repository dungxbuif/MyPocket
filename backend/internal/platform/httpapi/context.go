package httpapi

import "context"

type correlationKey struct{}

func withCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey{}, id)
}

func correlationID(ctx context.Context) string {
	id, _ := ctx.Value(correlationKey{}).(string)
	if id == "" {
		return "req_unavailable"
	}
	return id
}
