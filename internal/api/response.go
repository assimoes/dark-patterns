package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// decodeJSON reads the body into a T. on a bad body it writes a 400 and returns ok=false, so a
// handler just does: req, ok := decodeJSON[X](s, w, r); if !ok { return }.
func decodeJSON[T any](s *Server, w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad json body", err)
		return v, false
	}

	return v, true
}

func writeList[Row, Out any](
	s *Server,
	w http.ResponseWriter,
	r *http.Request,
	load func(context.Context) ([]Row, error),
	msg string,
	mapRow func(Row) Out,
) {
	rows, err := load(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, msg, err)
		return
	}

	out := make([]Out, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}

	s.writeJSON(w, http.StatusOK, out)
}

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
