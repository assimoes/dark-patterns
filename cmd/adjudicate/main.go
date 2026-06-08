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
		logger:  logger,
		auditor: auditor.ID,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/worklist", a.handleWorklist)
	mux.HandleFunc("GET /api/cells/{individual}/{pattern}", a.handleCell)
	mux.HandleFunc("POST /api/decisions", a.handleDecision)

	logger.Info("adjudication API up", "addr", *addr)

	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND-ORIGIN"))
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

func (a *api) handleWorklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	goldRun := atoi32(r.URL.Query().Get("gold_run"))
	perGame := atoi64(r.URL.Query().Get("per_game"))
	taxVer := atoi32(r.URL.Query().Get("tax_version"))

	gold, err := a.q.GetRun(ctx, goldRun)
	if err != nil {
		http.Error(w, "gold run not found", http.StatusNotFound)
		return
	}

	subset, err := a.q.SampleStratifiedIndividuals(ctx, db.SampleStratifiedIndividualsParams{
		PopulationID: gold.PopulationID,
		PerGame:      perGame,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	patterns, _ := a.q.GetMesoPatternCodes(ctx, taxVer)
	decided, _ := a.q.ListDedicedCells(ctx, goldRun)

	done := map[[2]int64]bool{}

	for _, d := range decided {
		done[[2]int64{d.IndividualID, int64(d.PatternID)}] = true
	}

	out := make([]worklistItem, 0, len(subset)*len(patterns))

	for _, s := range subset {
		for _, p := range patterns {
			out = append(out, worklistItem{
				IndividualID: s.IndividualID, PatternID: p.ID, Code: p.Code,
				Decided: done[[2]int64{s.IndividualID, int64(p.ID)}],
			})
		}
	}

	writeJSON(w, out)
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
