package middleware

import (
	"net/http"
	"strings"

	"learning-english/backend/internal/utils"
)

func RequestID(headerName string) Middleware {
	if strings.TrimSpace(headerName) == "" {
		headerName = "X-Request-Id"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get(headerName))
			if requestID == "" {
				requestID = utils.GenerateRequestID()
			}

			w.Header().Set(headerName, requestID)
			next.ServeHTTP(w, r.WithContext(utils.ContextWithRequestID(r.Context(), requestID)))
		})
	}
}
