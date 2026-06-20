package api

import (
	"log/slog"
	"net/http"
	"time"

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

	mux.HandleFunc("POST /api/images", s.createImages)

	mux.HandleFunc("PUT /api/games/{id}", s.updateGame)
	mux.HandleFunc("DELETE /api/games/{id}", s.deleteGame)
	mux.HandleFunc("GET /api/prompts/{id}", s.getPrompt)
	mux.HandleFunc("POST /api/prompts", s.createPrompt)
	mux.HandleFunc("PUT /api/prompts/{id}", s.updatePrompt)
	mux.HandleFunc("DELETE /api/prompts/{id}", s.deletePrompt)
	mux.HandleFunc("PUT /api/annotators/{id}", s.updateAnnotator)
	mux.HandleFunc("DELETE /api/annotators/{id}", s.deleteAnnotator)
	mux.HandleFunc("GET /api/populations/{id}/impact", s.populationImpact)
	mux.HandleFunc("DELETE /api/populations/{id}", s.deletePopulation)
	mux.HandleFunc("GET /api/runs/{id}/impact", s.runImpact)
	mux.HandleFunc("DELETE /api/runs/{id}", s.deleteRun)

	return s.withCORS(s.withLogging(mux))
}

// withCORS lets the configured dev origin call the API cross-origin and answers preflight OPTIONS.
func (s *Server) withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// statusRecorder remembers the status code a handler wrote so the logger can report it. http keeps it
// hidden once written, so we capture it on the way out.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withLogging prints one line per request: method, path, status and how long it took. handlers that
// never call WriteHeader still come out as 200, which is what net/http sends.
func (s *Server) withLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		h.ServeHTTP(rec, r)

		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", rec.status,
			"ms", time.Since(start).Milliseconds(),
		)
	})
}
