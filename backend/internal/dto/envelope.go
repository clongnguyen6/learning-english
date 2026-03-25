package dto

type APIResponse struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Error   *APIError      `json:"error"`
	Meta    map[string]any `json:"meta"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
