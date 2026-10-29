package httpapi

import "context"

type correlationKey struct{}
type authUserIDKey struct{}
type authMethodKey struct{}

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

func withAuthenticatedUser(ctx context.Context, userID string, method string) context.Context {
	ctx = context.WithValue(ctx, authUserIDKey{}, userID)
	return context.WithValue(ctx, authMethodKey{}, method)
}

func authenticatedUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(authUserIDKey{}).(string)
	return userID
}

func authenticatedMethod(ctx context.Context) string {
	method, _ := ctx.Value(authMethodKey{}).(string)
	return method
}
