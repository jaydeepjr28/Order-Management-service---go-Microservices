package common

import (
	"encoding/json"
	"net/http"
)

// WriteJSON is a helper function that marshals the provided data to JSON
// and writes it to the http.ResponseWriter with the specified HTTP status code.
// It automatically sets the Content-Type header to application/json.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ReadJSON is a helper function that decodes the JSON payload from the
// HTTP request body into the provided data structure.
// Returns an error if the decoding fails (e.g., malformed JSON or type mismatch).
func ReadJSON(r *http.Request, data any) error {
	return json.NewDecoder(r.Body).Decode(data)
}

// WriteError is a helper function that formats an error message as a JSON object
// ({"error": "message"}) and writes it to the response with the given status code.
// This ensures frontend clients receive a consistent error format.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}