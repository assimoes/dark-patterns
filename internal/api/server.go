package api

import (
	"log/slog"
	"net/http"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// Server holds the handler dependencies.
type Server struct {
	q           db.Querier
	pool        *pgxpool.Pool
	auditor     int32
	sameSite    http.SameSite
	riverClient *river.Client[pgx.Tx]
	logger      *slog.Logger
	origin      string
}

// NewServer wires the api server. origin is the browser origin allowed by CORS (the dev frontend).
func NewServer(pool *pgxpool.Pool, auditor int32, origin string,
	sameSite http.SameSite, riverClient *river.Client[pgx.Tx], logger *slog.Logger) *Server {
	return &Server{
		q:           db.New(pool),
		pool:        pool,
		auditor:     auditor,
		sameSite:    sameSite,
		riverClient: riverClient,
		logger:      logger,
		origin:      origin,
	}
}

// Routes returns the wired handler behind the CORS middleware.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/dashboard", s.dashboard)
	mux.HandleFunc("GET /api/games/{gameId}/populations", s.gamePopulations)
	mux.HandleFunc("GET /api/games/{gameId}/populations/{populationId}/models", s.gamePopulationModels)
	mux.HandleFunc("GET /api/games/{gameId}/runs/{runId}/models", s.gameRunModels)
	mux.HandleFunc("GET /api/populations/{populationId}/panel", s.populationPanel)
	mux.HandleFunc("POST /api/reviews/{reviewId}/decisions", s.reviewDecisions)

	mux.HandleFunc("POST /api/games", s.createGame)
	mux.HandleFunc("POST /api/annotators", s.createAnnotator)
	mux.HandleFunc("POST /api/populations", s.createPopulation)
	mux.HandleFunc("POST /api/runs", s.createRun)
	mux.HandleFunc("POST /api/scrapes", s.createScrape)
	mux.HandleFunc("POST /api/runs/{runID}/annotations", s.enqueueAnnotations)

	mux.HandleFunc("GET /api/populations", s.listPopulations)
	mux.HandleFunc("GET /api/populations/{populationId}", s.populationDetail)
	mux.HandleFunc("GET /api/annotators", s.listAnnotators)
	mux.HandleFunc("GET /api/prompts", s.listPrompts)
	mux.HandleFunc("GET /api/runs", s.listRuns)
	mux.HandleFunc("GET /api/runs/{runId}", s.runDetail)
	mux.HandleFunc("GET /api/games", s.listGames)

	mux.HandleFunc("POST /api/adjudication-samples", s.createAdjudicationSample)
	mux.HandleFunc("GET /api/runs/{runId}/adjudication-sample", s.runAdjudicationSample)
	mux.HandleFunc("GET /api/reviews/{reviewId}/adjudication", s.reviewAdjudication)
	mux.HandleFunc("GET /api/reviews/{reviewId}/blind", s.reviewBlind)

	return s.withCORS(mux)
}

// withCORS lets the configured dev origin call the API cross-origin and answers preflight OPTIONS.
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
