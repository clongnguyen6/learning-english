package handlers

import (
	"net/http"

	"learning-english/backend/internal/config"
)

type rootResponse struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Message     string `json:"message"`
}

func Root(cfg config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1" {
			NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet {
			MethodNotAllowed(w, r, http.MethodGet)
			return
		}

		WriteSuccess(w, http.StatusOK, rootResponse{
			Service:     "learning-english-api",
			Environment: cfg.App.Env,
			Message:     "api runtime contract initialized",
		}, requestMeta(r))
	})
}
