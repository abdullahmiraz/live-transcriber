package platform

import (
	"encoding/json"
	"net/http"
)

// APIError is the standard error envelope (see docs/api-design.md).
type APIError struct {
	Error APIErrorBody `json:"error"`
}

// APIErrorBody carries the machine code and human message.
type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON writes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// WriteError writes the standard error envelope.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, APIError{Error: APIErrorBody{Code: code, Message: message}})
}

// DecodeJSON decodes the request body into v, rejecting unknown fields.
func DecodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
