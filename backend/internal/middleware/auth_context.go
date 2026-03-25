package middleware

import "context"

type authContextKey string

const authenticatedUserContextKey authContextKey = "authenticated_user"

type AuthenticatedUser struct {
	UserID string
	Role   string
}

func ContextWithAuthenticatedUser(ctx context.Context, user AuthenticatedUser) context.Context {
	return context.WithValue(ctx, authenticatedUserContextKey, user)
}

func AuthenticatedUserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(authenticatedUserContextKey).(AuthenticatedUser)
	return user, ok
}
