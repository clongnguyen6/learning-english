package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"learning-english/backend/internal/dto"
	"learning-english/backend/internal/utils"
)

func WriteSuccess(w http.ResponseWriter, status int, data any, meta map[string]any) {
	writeJSON(w, status, dto.APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
		Meta:    normalizeMeta(meta),
	})
}

func WriteError(w http.ResponseWriter, status int, code, message string, details any, meta map[string]any) {
	writeJSON(w, status, dto.APIResponse{
		Success: false,
		Data:    nil,
		Error: &dto.APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: normalizeMeta(meta),
	})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(
		w,
		http.StatusNotFound,
		"NOT_FOUND",
		fmt.Sprintf("route %s %s not found", r.Method, r.URL.Path),
		nil,
		requestMeta(r),
	)
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request, allowedMethods ...string) {
	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	}

	WriteError(
		w,
		http.StatusMethodNotAllowed,
		"METHOD_NOT_ALLOWED",
		fmt.Sprintf("route %s %s does not allow this method", r.Method, r.URL.Path),
		map[string]any{
			"allowed_methods": allowedMethods,
		},
		requestMeta(r),
	)
}

func requestMeta(r *http.Request) map[string]any {
	meta := map[string]any{}
	if requestID := utils.RequestIDFromContext(r.Context()); requestID != "" {
		meta["request_id"] = requestID
	}

	return meta
}

func writeJSON(w http.ResponseWriter, status int, response dto.APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(response)
}

func normalizeMeta(meta map[string]any) map[string]any {
	if meta == nil {
		return map[string]any{}
	}

	return meta
}
