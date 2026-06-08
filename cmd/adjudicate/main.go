package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/assimoes/dsr/internal/adjudicate"
	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type api struct {
	q       *db.Queries
	pool    *pgxpool.Pool
	logger  *slog.Logger
	auditor int32
}

type voteDTO struct {
	ModelSlug   string `json:"model_slug"`
	Present     bool   `json:"present"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
}

type cellDTO struct {
	IndividualID int64     `json:"individual_id"`
	PatternID    int32     `json:"pattern_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	ReviewText   string    `json:"review_text"`
	PanelVote    bool      `json:"panel_vote"`
	NPresent     int       `json:"n_present"`
	NTotal       int       `json:"n_total"`
	Votes        []voteDTO `json:"votes"`
}

type worklistItem struct {
	IndividualID int64  `json:"individual_id"`
	PatternID    int32  `json:"pattern_id"`
	Code         string `json:"code"`
	Decided      bool   `json:"declined"`
}

type decisionReq struct {
	GoldRun      int32 `json:"gold_run"`
	PanelRun     int32 `json:"panel_run"`
	IndividualID int64 `json:"individual_id"`
	PatternID    int32 `json:"pattern_id"`
	Label        bool  `json:"label"`
}

type reviewItem struct {
	IndividualID   int64 `json:"individual_id"`
	ExternalGameID int32 `json:"external_game_id"`
	Decided        int   `json:"decided"`
	Total          int   `json:"total"`
}

type detectionDTO struct {
	ModelSlug   string `json:"model_slug"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
}

type reviewPatternDTO struct {
	PatternID   int32          `json:"pattern_id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Family      string         `json:"family"`
	Description string         `json:"description"`
	PanelVote   bool           `json:"panel_vote"`
	NPresent    int            `json:"n_present"`
	NTotal      int            `json:"n_total"`
	Decided     bool           `json:"decided"`
	FinalLabel  bool           `json:"final_label"`
	Detections  []detectionDTO `json:"detections"`
}

type reviewDTO struct {
	IndividualID int64              `json:"individual_id"`
	ReviewText   string             `json:"review_text"`
	NTotal       int                `json:"n_total"`
	Patterns     []reviewPatternDTO `json:"patterns"`
}

type reviewDecisionsReq struct {
	GoldRun   int32 `json:"gold_run"`
	PanelRun  int32 `json:"panel_run"`
	Decisions []struct {
		PatternID int32 `json:"pattern_id"`
		Label     bool  `json:"label"`
	} `json:"decisions"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	addr := flag.String("addr", ":8080", "listen address for the adjudication API")
	adjLabel := flag.String("adjudicator", "author", "label of the kind=human annotator")

	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL required")
		os.Exit(2)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := db.New(pool)

	auditor, err := q.GetAnnotatorByLabel(ctx, *adjLabel)
	if err != nil || auditor.Kind != "human" {
		logger.Error("adjudicator must be an existing kind=human annotator", "label", *adjLabel, "err", err)
		os.Exit(1)
	}

	a := &api{
		q:       q,
		pool:    pool,
		logger:  logger,
		auditor: auditor.ID,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/worklist", a.handleWorklist)
	mux.HandleFunc("GET /api/reviews/{individual}", a.handleReview)
	mux.HandleFunc("POST /api/reviews/{individual}/decisions", a.handleReviewDecisions)
	mux.HandleFunc("GET /api/cells/{individual}/{pattern}", a.handleCell)
	mux.HandleFunc("POST /api/decisions", a.handleDecision)

	logger.Info("adjudication API up", "addr", *addr)

	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}

func withCORS(h http.Handler) http.Handler {
	origin := os.Getenv("FRONTEND_ORIGIN")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// handleWorklist lists the reviews in scope (one row per review), with how many of its patterns are
// already adjudicated, so the auditor can see progress and pick what to work on.
func (a *api) handleWorklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	goldRun := atoi32(r.URL.Query().Get("gold_run"))
	panelRun := atoi32(r.URL.Query().Get("panel_run"))
	perGame := atoi64(r.URL.Query().Get("per_game"))
	taxVer := atoi32(r.URL.Query().Get("tax_version"))

	gold, err := a.q.GetRun(ctx, goldRun)
	if err != nil {
		http.Error(w, "gold run not found", http.StatusNotFound)
		return
	}

	subset, err := a.q.SampleStratifiedIndividuals(ctx, db.SampleStratifiedIndividualsParams{
		PopulationID: gold.PopulationID,
		PanelRunID:   panelRun,
		PerGame:      perGame,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	patterns, _ := a.q.GetMesoPatternCodes(ctx, taxVer)
	total := len(patterns)

	counts, _ := a.q.CountAdjudicationsPerReview(ctx, goldRun)
	decided := map[int64]int{}

	for _, c := range counts {
		decided[c.IndividualID] = int(c.Decided)
	}

	out := make([]reviewItem, 0, len(subset))

	for _, s := range subset {
		out = append(out, reviewItem{
			IndividualID:   s.IndividualID,
			ExternalGameID: s.ExternalGameID,
			Decided:        decided[s.IndividualID],
			Total:          total,
		})
	}

	writeJSON(w, out)
}

// handleReview returns one review's text plus every pattern in the codebook: its definition, the
// panel's vote, and — for the ones a model flagged — who flagged it with their evidence/explanation.
func (a *api) handleReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ind, _ := strconv.ParseInt(r.PathValue("individual"), 10, 64)
	panelRun := atoi32(r.URL.Query().Get("panel_run"))
	goldRun := atoi32(r.URL.Query().Get("gold_run"))
	taxVer := atoi32(r.URL.Query().Get("tax_version"))

	if taxVer == 0 {
		taxVer = 1
	}

	text, _ := a.q.GetReviewText(ctx, ind)

	tax, err := a.q.ListTaxonomy(ctx, taxVer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nTotal, _ := a.q.CountCompletedRaters(ctx, db.CountCompletedRatersParams{
		PanelRunID: panelRun, IndividualID: ind,
	})

	dets, _ := a.q.ListReviewDetections(ctx, db.ListReviewDetectionsParams{
		PanelRunID: panelRun, IndividualID: ind,
	})

	byPattern := map[int32][]detectionDTO{}

	for _, d := range dets {
		byPattern[d.PatternID] = append(byPattern[d.PatternID], detectionDTO{
			ModelSlug:   d.ModelSlug,
			Evidence:    d.Evidence,
			Explanation: d.Explanation,
		})
	}

	adjs, _ := a.q.ListReviewAdjudications(ctx, db.ListReviewAdjudicationsParams{
		RunID: goldRun, IndividualID: ind,
	})

	gold := map[int32]bool{}

	for _, ad := range adjs {
		gold[ad.PatternID] = ad.FinalLabel
	}

	patterns := make([]reviewPatternDTO, 0, len(tax))

	for _, t := range tax {
		hits := byPattern[t.ID]
		if hits == nil {
			hits = []detectionDTO{}
		}

		final, decided := gold[t.ID]

		patterns = append(patterns, reviewPatternDTO{
			PatternID:   t.ID,
			Code:        t.Code,
			Name:        t.Name,
			Family:      t.Family,
			Description: t.Description,
			PanelVote:   adjudicate.Majority(len(hits), int(nTotal)),
			NPresent:    len(hits),
			NTotal:      int(nTotal),
			Decided:     decided,
			FinalLabel:  final,
			Detections:  hits,
		})
	}

	writeJSON(w, reviewDTO{
		IndividualID: ind,
		ReviewText:   text,
		NTotal:       int(nTotal),
		Patterns:     patterns,
	})
}

// handleReviewDecisions writes every pattern's decision for one review in a single transaction. The
// panel seed is rebuilt and frozen per pattern from a fresh trusted read, exactly as the single-cell
// path does, so the browser only ever sends the labels.
func (a *api) handleReviewDecisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ind, _ := strconv.ParseInt(r.PathValue("individual"), 10, 64)

	var req reviewDecisionsReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	for _, d := range req.Decisions {
		counts, err := q.GetPanelVoteForCell(ctx, db.GetPanelVoteForCellParams{
			PanelRunID: req.PanelRun, IndividualID: ind, PatternID: d.PatternID,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		perRater, _ := q.ListPanelVotesForCell(ctx, db.ListPanelVotesForCellParams{
			PanelRunID: req.PanelRun, IndividualID: ind, PatternID: d.PatternID,
		})

		vote := adjudicate.Majority(int(counts.NPresent), int(counts.NTotal))
		votes := make([]adjudicate.PanelVerdict, 0, len(perRater))

		for _, v := range perRater {
			votes = append(votes, adjudicate.PanelVerdict{
				AnnotatorID: v.AnnotatorID,
				ModelSlug:   v.ModelSlug,
				Present:     v.Present,
				Evidence:    v.Evidence,
				Explanation: v.Explanation,
			})
		}

		seed := adjudicate.PanelSeed{
			Vote:     vote,
			NPresent: int(counts.NPresent),
			NTotal:   int(counts.NTotal),
			Votes:    votes,
		}

		direction := "replacement"

		if d.Label == vote {
			direction = "confirmation"
		}

		if err := q.UpsertAdjudication(ctx, db.UpsertAdjudicationParams{
			RunID:                   req.GoldRun,
			IndividualID:            ind,
			PatternID:               d.PatternID,
			FinalLabel:              d.Label,
			Direction:               direction,
			AdjudicatorID:           a.auditor,
			PanelSeedAtAdjudication: seed.JSON(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{"status": "ok", "saved": len(req.Decisions)})
}

func (a *api) handleCell(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ind, _ := strconv.ParseInt(r.PathValue("individual"), 10, 64)
	pat := atoi32(r.PathValue("pattern"))

	panelRun := atoi32(r.URL.Query().Get("panel_run"))

	text, _ := a.q.GetReviewText(ctx, ind)
	counts, err := a.q.GetPanelVoteForCell(ctx, db.GetPanelVoteForCellParams{
		PanelRunID: panelRun, IndividualID: ind, PatternID: pat,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	perRater, _ := a.q.ListPanelVotesForCell(ctx, db.ListPanelVotesForCellParams{
		PanelRunID: panelRun, IndividualID: ind, PatternID: pat,
	})

	codes, _ := a.q.GetMesoPatternCodes(ctx, 1)

	code, name := codeName(codes, pat)

	votes := make([]voteDTO, 0, len(perRater))

	for _, v := range perRater {
		votes = append(votes, voteDTO{
			ModelSlug:   v.ModelSlug,
			Present:     v.Present,
			Evidence:    v.Evidence,
			Explanation: v.Explanation,
		})
	}

	writeJSON(w, cellDTO{
		IndividualID: ind,
		PatternID:    pat,
		Code:         code,
		Name:         name,
		ReviewText:   text,
		PanelVote:    adjudicate.Majority(int(counts.NPresent), int(counts.NTotal)),
		NPresent:     int(counts.NPresent),
		NTotal:       int(counts.NTotal),
		Votes:        votes,
	})
}

func (a *api) handleDecision(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req decisionReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	counts, err := a.q.GetPanelVoteForCell(ctx, db.GetPanelVoteForCellParams{
		PanelRunID:   req.PanelRun,
		IndividualID: req.IndividualID,
		PatternID:    req.PatternID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	perRater, _ := a.q.ListPanelVotesForCell(ctx, db.ListPanelVotesForCellParams{
		PanelRunID:   req.PanelRun,
		IndividualID: req.IndividualID,
		PatternID:    req.PatternID,
	})

	vote := adjudicate.Majority(int(counts.NPresent), int(counts.NTotal))
	votes := make([]adjudicate.PanelVerdict, 0, len(perRater))

	for _, v := range perRater {
		votes = append(votes, adjudicate.PanelVerdict{
			AnnotatorID: v.AnnotatorID,
			ModelSlug:   v.ModelSlug,
			Present:     v.Present,
			Evidence:    v.Evidence,
			Explanation: v.Explanation,
		})
	}

	seed := adjudicate.PanelSeed{
		Vote:     vote,
		NPresent: int(counts.NPresent),
		NTotal:   int(counts.NTotal),
		Votes:    votes,
	}

	direction := "replacement"

	if req.Label == vote {
		direction = "confirmation"
	}

	if err := a.q.UpsertAdjudication(ctx, db.UpsertAdjudicationParams{
		RunID:                   req.GoldRun,
		IndividualID:            req.IndividualID,
		PatternID:               req.PatternID,
		FinalLabel:              req.Label,
		Direction:               direction,
		AdjudicatorID:           a.auditor,
		PanelSeedAtAdjudication: seed.JSON(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok", "direction": direction})
}

func atoi32(s string) int32 {
	r, _ := strconv.Atoi(s)

	return int32(r)
}

func atoi64(s string) int64 {
	r, _ := strconv.ParseInt(s, 10, 64)

	return r
}

func codeName(codes []db.GetMesoPatternCodesRow, id int32) (code, name string) {
	for _, c := range codes {
		if c.ID == id {
			return c.Code, c.Name
		}
	}

	return "", ""
}
