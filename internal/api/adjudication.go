package api

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5"
)

// strataKey groups the candidate pool by the two axes we draw within: each (game, stratum) cell is
// sampled independently so every game and every stratum is represented in proportion to its target.
type strataKey struct {
	gameID  int32
	stratum string
}

// createAdjudicationSample draws and persists a stratified sample. It resolves the panel run's
// population and its gold run (creating the gold run if the population has none yet), classifies every
// completed review into one stratum, then for each (game x stratum) cell shuffles with a seeded RNG and
// takes min(target, cell size). The whole sample (header + one item per selected review, each carrying
// its stratum and selection probability) is written in one transaction so a half-drawn sample is never
// observable. The seed is stored so the draw is reproducible and auditable.
func (s *Server) createAdjudicationSample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad json body", err)
		return
	}

	if req.PanelRunID == 0 {
		s.writeError(w, http.StatusBadRequest, "panelRunId is required", nil)
		return
	}

	panelRun, err := s.q.GetRun(ctx, req.PanelRunID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "panel run not found", err)
		return
	}

	// Resolve the gold run for the population, or create one mirroring the panel run's configuration.
	goldRun, err := s.resolveGoldRun(ctx, panelRun)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "resolve gold run", err)
		return
	}

	rows, err := s.q.ClassifyReviewsForSampling(ctx, db.ClassifyReviewsForSamplingParams{
		PanelRunID:   req.PanelRunID,
		PopulationID: panelRun.PopulationID,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "classify reviews", err)
		return
	}

	// Group the candidate pool into (game x stratum) cells.
	groups := make(map[strataKey][]db.ClassifyReviewsForSamplingRow)
	for _, row := range rows {
		k := strataKey{gameID: row.ExternalGameID, stratum: row.Stratum}
		groups[k] = append(groups[k], row)
	}

	target := map[string]int{
		"flagged_majority": req.FlaggedMajority,
		"flagged_split":    req.FlaggedSplit,
		"silent":           req.Silent,
	}

	// Seed: use the supplied one for a reproducible draw, otherwise generate and store one so the
	// persisted sample is still reproducible after the fact.
	seed := time.Now().UnixNano()
	if req.Seed != nil {
		seed = *req.Seed
	}
	rng := rand.New(rand.NewSource(seed))

	params, err := json.Marshal(map[string]int{
		"flagged_majority": req.FlaggedMajority,
		"flagged_split":    req.FlaggedSplit,
		"silent":           req.Silent,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "marshal params", err)
		return
	}

	// Draw the selected items before opening the transaction. Iterating the cells in a stable key
	// order keeps the draw deterministic for a given seed regardless of map iteration order.
	keys := make([]strataKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].gameID != keys[j].gameID {
			return keys[i].gameID < keys[j].gameID
		}
		return keys[i].stratum < keys[j].stratum
	})

	type selected struct {
		row  db.ClassifyReviewsForSamplingRow
		prob float64
	}

	var picks []selected
	counts := stratumCounts{}

	for _, k := range keys {
		cell := groups[k]
		n := target[k.stratum]
		if n > len(cell) {
			n = len(cell)
		}
		if n <= 0 {
			continue
		}

		// Shuffle a copy with the seeded RNG, then take the first n.
		shuffled := make([]db.ClassifyReviewsForSamplingRow, len(cell))
		copy(shuffled, cell)
		rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		prob := float64(n) / float64(len(cell))
		for _, row := range shuffled[:n] {
			picks = append(picks, selected{row: row, prob: prob})
		}

		switch k.stratum {
		case "flagged_majority":
			counts.FlaggedMajority += n
		case "flagged_split":
			counts.FlaggedSplit += n
		case "silent":
			counts.Silent += n
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	sampleID, err := q.InsertAdjudicationSample(ctx, db.InsertAdjudicationSampleParams{
		PanelRunID: req.PanelRunID,
		GoldRunID:  goldRun,
		Strategy:   "stratified",
		Seed:       seed,
		Params:     params,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "insert sample", err)
		return
	}

	for _, p := range picks {
		if err := q.InsertAdjudicationSampleItem(ctx, db.InsertAdjudicationSampleItemParams{
			SampleID:       sampleID,
			IndividualID:   p.row.IndividualID,
			ExternalGameID: p.row.ExternalGameID,
			Stratum:        p.row.Stratum,
			SelectionProb:  p.prob,
		}); err != nil {
			s.writeError(w, http.StatusInternalServerError, "insert sample item", err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit tx", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, createSampleResponse{
		SampleID:  sampleID,
		GoldRunID: goldRun,
		Seed:      seed,
		Counts:    counts,
	})
}

// resolveGoldRun returns the population's gold run, creating one that mirrors the panel run's
// population, prompt, taxonomy version and temperature when none exists yet. There is one gold run per
// population; samples and decisions for the same population all land on the same run.
func (s *Server) resolveGoldRun(ctx context.Context, panelRun db.Run) (int32, error) {
	goldRun, err := s.q.GetGoldRunForPopulation(ctx, panelRun.PopulationID)
	if err == nil {
		return goldRun, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	// No gold run yet: create one from the panel run's configuration.
	return s.q.CreateRun(ctx, db.CreateRunParams{
		RunType:         "gold",
		PopulationID:    panelRun.PopulationID,
		PromptID:        panelRun.PromptID,
		Temperature:     panelRun.Temperature,
		TopP:            panelRun.TopP,
		Params:          nil,
		TaxonomyVersion: panelRun.TaxonomyVersion,
		AnnotatorIds:    nil,
	})
}

// runAdjudicationSample returns the persisted sample for a panel run and its frozen review queue. The
// queue carries each review's stratum and how many of its patterns already have a gold label, so the
// frontend can show progress. 404 when the run has no sample yet.
func (s *Server) runAdjudicationSample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	runID, err := parseInt32(r.PathValue("runId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid runId", err)
		return
	}

	sample, err := s.q.GetLatestSampleForRun(ctx, runID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.writeError(w, http.StatusNotFound, "no sample for this run", err)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "load sample", err)
		return
	}

	rows, err := s.q.ListSampleReviews(ctx, sample.ID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load sample reviews", err)
		return
	}

	reviews := make([]SampleReview, 0, len(rows))
	for _, rv := range rows {
		reviews = append(reviews, SampleReview{
			ID:       strconv.FormatInt(rv.IndividualID, 10),
			GameID:   gameID(rv.ExternalGameID),
			Stratum:  rv.Stratum,
			VotedUp:  rv.VotedUp,
			Language: rv.Lang,
			Decided:  int(rv.Decided),
		})
	}

	s.writeJSON(w, http.StatusOK, sampleResponse{
		SampleID:  sample.ID,
		GoldRunID: sample.GoldRunID,
		Reviews:   reviews,
	})
}

// reviewAdjudication returns the full review the auditor decides on: the body and metadata, the panel
// models, the per-model detections (a detection row IS a Present vote; Absent otherwise — the frontend
// reconstructs every model's per-pattern vote from detections + panelModels), and any existing gold
// labels keyed by pattern code. The {reviewId} path value is the individual id; the panel run comes
// from ?run= or, when absent, the review's most recent panel run.
func (s *Server) reviewAdjudication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	individualID, err := strconv.ParseInt(r.PathValue("reviewId"), 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid reviewId", err)
		return
	}

	panelRun, err := s.resolvePanelRun(ctx, r, individualID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "no panel run for this review", err)
		return
	}

	meta, err := s.q.GetReviewMeta(ctx, individualID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "review not found", err)
		return
	}

	run, err := s.q.GetRun(ctx, panelRun)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load panel run", err)
		return
	}

	taxVersion := int32(1)
	if run.TaxonomyVersion != nil {
		taxVersion = *run.TaxonomyVersion
	}

	codes, err := s.q.GetMesoPatternCodes(ctx, taxVersion)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load taxonomy codes", err)
		return
	}
	codeByID := make(map[int32]string, len(codes))
	for _, c := range codes {
		codeByID[c.ID] = c.Code
	}

	dets, err := s.q.ListReviewDetections(ctx, db.ListReviewDetectionsParams{
		PanelRunID:   panelRun,
		IndividualID: individualID,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load detections", err)
		return
	}

	detections := make([]Detection, 0, len(dets))
	for _, d := range dets {
		detections = append(detections, Detection{
			PatternCode: codeByID[d.PatternID],
			Model:       d.ModelSlug,
			Evidence:    d.Evidence,
			Explanation: d.Explanation,
		})
	}

	// The panel model list lets the frontend infer Absent (a model with no detection row on a pattern).
	annotators, err := s.q.ListRunAnnotators(ctx, panelRun)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load panel models", err)
		return
	}
	panelModels := make([]string, 0, len(annotators))
	for _, a := range annotators {
		panelModels = append(panelModels, a.ModelSlug)
	}

	// Existing gold labels for the auditor to revisit, keyed by pattern code for the frontend.
	goldRun, err := s.q.GetGoldRunForPopulation(ctx, run.PopulationID)
	goldLabels := map[string]bool{}
	if err == nil {
		adjs, err := s.q.ListReviewAdjudications(ctx, db.ListReviewAdjudicationsParams{
			RunID:        goldRun,
			IndividualID: individualID,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "load gold labels", err)
			return
		}
		for _, a := range adjs {
			if code, ok := codeByID[a.PatternID]; ok {
				goldLabels[code] = a.FinalLabel
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		s.writeError(w, http.StatusInternalServerError, "resolve gold run", err)
		return
	}

	s.writeJSON(w, http.StatusOK, adjudicationReview{
		ID:          strconv.FormatInt(individualID, 10),
		GameID:      gameID(meta.ExternalGameID),
		VotedUp:     meta.VotedUp,
		Language:    meta.Lang,
		Body:        meta.Body,
		PanelModels: panelModels,
		Detections:  detections,
		GoldLabels:  goldLabels,
	})
}

// reviewBlind returns the panel-FREE projection for the blind pass: the review body and metadata and
// nothing about the panel. This is the provably-blind boundary — no detections, no votes, no gold
// labels can leak through it. The {reviewId} path value is the individual id.
func (s *Server) reviewBlind(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	individualID, err := strconv.ParseInt(r.PathValue("reviewId"), 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid reviewId", err)
		return
	}

	meta, err := s.q.GetReviewMeta(ctx, individualID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "review not found", err)
		return
	}

	s.writeJSON(w, http.StatusOK, blindReview{
		ID:       strconv.FormatInt(individualID, 10),
		GameID:   gameID(meta.ExternalGameID),
		VotedUp:  meta.VotedUp,
		Language: meta.Lang,
		Body:     meta.Body,
	})
}

// resolvePanelRun picks the panel run for a review: the ?run= query param when present and valid,
// otherwise the review's most recent llm_panel run.
func (s *Server) resolvePanelRun(ctx context.Context, r *http.Request, individualID int64) (int32, error) {
	if q := r.URL.Query().Get("run"); q != "" {
		return parseInt32(q)
	}

	return s.q.GetPanelRunForReview(ctx, individualID)
}
