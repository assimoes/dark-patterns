package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON encodes v as the JSON response body. Encode errors are logged, not surfaced.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("encode response", "err", err)
	}
}

// writeError sends a plain-text error and logs the cause.
func (s *Server) writeError(w http.ResponseWriter, status int, msg string, err error) {
	if err != nil {
		s.logger.Error(msg, "err", err)
	}

	http.Error(w, msg, status)
}
