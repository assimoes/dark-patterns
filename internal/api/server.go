package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds the dependencies the handlers need. q is the generated query interface for reads;
// pool is kept for the one write path that spans several statements in a transaction; auditor is the
// id of the kind=human annotator that decisions are attributed to.
type Server struct {
	q       db.Querier
	pool    *pgxpool.Pool
	auditor int32
	logger  *slog.Logger
	origin  string
}

// NewServer wires the api server. origin is the browser origin allowed by CORS (the dev frontend).
func NewServer(pool *pgxpool.Pool, auditor int32, origin string, logger *slog.Logger) *Server {
	return &Server{
		q:       db.New(pool),
		pool:    pool,
		auditor: auditor,
		logger:  logger,
		origin:  origin,
	}
}

// Routes returns the fully wired handler: the four endpoints behind the CORS middleware. Go 1.22+
// method+pattern routing means each route states its verb, and path wildcards are read with
// r.PathValue.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/dashboard", s.dashboard)
	mux.HandleFunc("GET /api/games/{gameId}/populations", s.gamePopulations)
	mux.HandleFunc("GET /api/games/{gameId}/populations/{populationId}/models", s.gamePopulationModels)
	mux.HandleFunc("GET /api/populations/{populationId}/panel", s.populationPanel)
	mux.HandleFunc("GET /api/runs/{runId}/reviews", s.runReviews)
	mux.HandleFunc("POST /api/reviews/{reviewId}/decisions", s.reviewDecisions)

	return s.withCORS(mux)
}

// withCORS allows the browser dev origin to call the API cross-origin and answers preflight
// OPTIONS requests. The allowed origin is configured by the composition root.
func (s *Server) withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// writeJSON encodes v as the response body. A failure to encode is logged, not surfaced, because the
// status line is already committed by the time encoding runs.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("encode response", "err", err)
	}
}

// writeError sends a plain-text error with the given status and logs the cause.
func (s *Server) writeError(w http.ResponseWriter, status int, msg string, err error) {
	if err != nil {
		s.logger.Error(msg, "err", err)
	}

	http.Error(w, msg, status)
}
