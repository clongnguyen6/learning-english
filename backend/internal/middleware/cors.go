package middleware

import "net/http"

func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, allowedOrigins) {
				headers := w.Header()
				headers.Add("Vary", "Origin")
				headers.Add("Vary", "Access-Control-Request-Method")
				headers.Add("Vary", "Access-Control-Request-Headers")
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Set("Access-Control-Allow-Credentials", "true")
				headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, If-Match, If-None-Match, X-Request-Id")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}
