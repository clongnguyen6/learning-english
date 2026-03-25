package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

func GenerateRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(bytes[:])
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}

	return value
}
