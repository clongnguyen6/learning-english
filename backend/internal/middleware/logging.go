package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"learning-english/backend/internal/utils"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(data)
	r.bytes += n

	return n, err
}

func Logging(logger *slog.Logger, requestIDHeader string) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &statusRecorder{ResponseWriter: w}
			startedAt := time.Now()

			next.ServeHTTP(recorder, r)

			if recorder.status == 0 {
				recorder.status = http.StatusOK
			}

			logger.Info(
				"http request complete",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"bytes", recorder.bytes,
				"duration_ms", time.Since(startedAt).Milliseconds(),
				"request_id", utils.RequestIDFromContext(r.Context()),
				"request_id_header", requestIDHeader,
			)
		})
	}
}
