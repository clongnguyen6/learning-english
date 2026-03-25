package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"learning-english/backend/internal/dto"
	"learning-english/backend/internal/utils"
)

func Recovery(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "panic", recovered, "request_id", utils.RequestIDFromContext(r.Context()))

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)

					response := dto.APIResponse{
						Success: false,
						Data:    nil,
						Error: &dto.APIError{
							Code:    "INTERNAL_SERVER_ERROR",
							Message: "internal server error",
						},
						Meta: map[string]any{
							"request_id": utils.RequestIDFromContext(r.Context()),
						},
					}

					encoder := json.NewEncoder(w)
					encoder.SetEscapeHTML(false)
					_ = encoder.Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
