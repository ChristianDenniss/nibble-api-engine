// Package session carries the signed-in account on the request context so
// domain modules can read it without importing the auth module.
package session

import "context"

type ctxKey struct{}

func WithAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, accountID)
}

// AccountID returns the signed-in account, or "" for guests.
func AccountID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
