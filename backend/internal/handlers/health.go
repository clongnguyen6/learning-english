package handlers

import (
	"context"
	"net/http"
)

type readinessPinger interface {
	Ping(context.Context) error
}

type HealthHandler struct {
	serviceName string
	version     string
	pinger      readinessPinger
}

func NewHealthHandler(serviceName, version string, pinger readinessPinger) HealthHandler {
	return HealthHandler{
		serviceName: serviceName,
		version:     version,
		pinger:      pinger,
	}
}

func (h HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": h.serviceName,
		"version": h.version,
	}, requestMeta(r))
}

func (h HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.pinger == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "readiness checks are not configured", nil, requestMeta(r))
		return
	}

	if err := h.pinger.Ping(r.Context()); err != nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "readiness checks failed", map[string]any{
			"dependency": "database",
			"cause":      err.Error(),
		}, requestMeta(r))
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]any{
		"status":  "ready",
		"service": h.serviceName,
		"version": h.version,
		"checks": map[string]any{
			"database": "ok",
		},
	}, requestMeta(r))
}
