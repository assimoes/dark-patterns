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

	"github.com/assimoes/dsr/internal/adjudicate"
	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5"
)

// strataKey is one (game, stratum) cell; each cell is sampled independently.
type strataKey struct {
	gameID  int32
	stratum string
}

// createAdjudicationSample draws and persists a stratified sample.
func (s *Server) createAdjudicationSample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.CreateSampleRequest
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

	// group into (game x stratum) cells
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

	// store the seed (supplied or generated) so the draw stays reproducible after the fact
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

	// draw before opening the tx; stable key order keeps the draw deterministic per seed
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
	counts := dto.StratumCounts{}

	for _, k := range keys {
		cell := groups[k]
		n := target[k.stratum]
		if n > len(cell) {
			n = len(cell)
		}
		if n <= 0 {
			continue
		}

		// shuffle a copy with the seeded RNG, take first n
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

	s.writeJSON(w, http.StatusCreated, dto.CreateSampleResponse{
		SampleID:  sampleID,
		GoldRunID: goldRun,
		Seed:      seed,
		Counts:    counts,
	})
}

// resolveGoldRun returns the populations gold run, creating one from the panel runs config if theres none
// yet. one gold run per population, so samples and decisions all land on the same run.
func (s *Server) resolveGoldRun(ctx context.Context, panelRun db.Run) (int32, error) {
	goldRun, err := s.q.GetGoldRunForPopulation(ctx, panelRun.PopulationID)
	if err == nil {
		return goldRun, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

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

// runAdjudicationSample returns the persisted sample for a panel run and its frozen review queue.
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

	reviews := make([]dto.SampleReview, 0, len(rows))
	for _, rv := range rows {
		reviews = append(reviews, dto.SampleReview{
			ID:       strconv.FormatInt(rv.IndividualID, 10),
			GameID:   gameID(rv.ExternalGameID),
			Stratum:  rv.Stratum,
			VotedUp:  rv.VotedUp,
			Language: rv.Lang,
			Decided:  int(rv.Decided),
		})
	}

	s.writeJSON(w, http.StatusOK, dto.SampleResponse{
		SampleID:  sample.ID,
		GoldRunID: sample.GoldRunID,
		Reviews:   reviews,
	})
}

// reviewAdjudication returns the full review the auditor decides on.
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

	detections := make([]dto.Detection, 0, len(dets))
	for _, d := range dets {
		detections = append(detections, dto.Detection{
			PatternCode: codeByID[d.PatternID],
			Model:       d.ModelSlug,
			Evidence:    d.Evidence,
			Explanation: d.Explanation,
		})
	}

	// panel model list lets the frontend infer Absent (model with no detection row on a pattern)
	annotators, err := s.q.ListRunAnnotators(ctx, panelRun)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load panel models", err)
		return
	}
	panelModels := make([]string, 0, len(annotators))
	for _, a := range annotators {
		panelModels = append(panelModels, a.ModelSlug)
	}

	// existing gold labels to revisit, keyed by pattern code for the frontend
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

	s.writeJSON(w, http.StatusOK, dto.AdjudicationReview{
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

// reviewBlind returns the panel-free projection for the blind pass.
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

	s.writeJSON(w, http.StatusOK, dto.BlindReview{
		ID:       strconv.FormatInt(individualID, 10),
		GameID:   gameID(meta.ExternalGameID),
		VotedUp:  meta.VotedUp,
		Language: meta.Lang,
		Body:     meta.Body,
	})
}

// resolvePanelRun picks the panel run for a review: ?run= if present, else the reviews newest llm_panel run.
func (s *Server) resolvePanelRun(ctx context.Context, r *http.Request, individualID int64) (int32, error) {
	if q := r.URL.Query().Get("run"); q != "" {
		return parseInt32(q)
	}

	return s.q.GetPanelRunForReview(ctx, individualID)
}

// reviewDecisions records the auditors present/absent calls for one review ({reviewId} is the
// individual id). each call freezes the panel state at decision time, the whole set is one tx so theres no
// partial save. responds 204.
func (s *Server) reviewDecisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	individualID, err := strconv.ParseInt(r.PathValue("reviewId"), 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid reviewId", err)
		return
	}

	var req dto.DecisionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad json body", err)
		return
	}

	// derive gold run and taxonomy version from the ?run= the screen sends, not "newest panel run":
	// a population can hold panel runs of different tax versions, and the read maps labels by this
	// runs pattern-version ids, so a mismatch here hides the saved labels on revisit
	panelRun, err := s.resolvePanelRun(ctx, r, individualID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "no panel run for this review", err)
		return
	}

	panelRunRow, err := s.q.GetRun(ctx, panelRun)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load panel run", err)
		return
	}

	taxVersion := int32(1)
	if panelRunRow.TaxonomyVersion != nil {
		taxVersion = *panelRunRow.TaxonomyVersion
	}

	goldRun, err := s.q.GetGoldRunForPopulation(ctx, panelRunRow.PopulationID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "no gold run for this review", err)
		return
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "begin tx", err)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	for code, label := range req.Decisions {
		final, ok := parseLabel(label)
		if !ok {
			s.writeError(w, http.StatusBadRequest, "decision must be 'present' or 'absent'", nil)
			return
		}

		patternID, err := q.GetPatternIDByCode(ctx, db.GetPatternIDByCodeParams{Code: code, Version: taxVersion})
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "unknown pattern code "+code, err)
			return
		}

		seed, vote, err := s.buildSeed(ctx, q, panelRun, individualID, patternID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "build panel seed", err)
			return
		}

		direction := "replacement"
		if final == vote {
			direction = "confirmation"
		}

		if err := q.UpsertAdjudication(ctx, db.UpsertAdjudicationParams{
			RunID:                   goldRun,
			IndividualID:            individualID,
			PatternID:               patternID,
			FinalLabel:              final,
			Direction:               direction,
			AdjudicatorID:           s.auditor,
			PanelSeedAtAdjudication: seed.JSON(),
		}); err != nil {
			s.writeError(w, http.StatusInternalServerError, "save adjudication", err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.writeError(w, http.StatusInternalServerError, "commit tx", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// buildSeed reconstructs the panel verdict for one (panel run, review, pattern) and returns the frozen
// seed plus the majority vote, so a decision can be marked confirmation or replacement.
func (s *Server) buildSeed(ctx context.Context, q db.Querier, panelRun int32, individualID int64, patternID int32) (adjudicate.PanelSeed, bool, error) {
	counts, err := q.GetPanelVoteForCell(ctx, db.GetPanelVoteForCellParams{
		PanelRunID:   panelRun,
		IndividualID: individualID,
		PatternID:    patternID,
	})
	if err != nil {
		return adjudicate.PanelSeed{}, false, err
	}

	perRater, err := q.ListPanelVotesForCell(ctx, db.ListPanelVotesForCellParams{
		PanelRunID:   panelRun,
		IndividualID: individualID,
		PatternID:    patternID,
	})
	if err != nil {
		return adjudicate.PanelSeed{}, false, err
	}

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

	return adjudicate.PanelSeed{
		Vote:     vote,
		NPresent: int(counts.NPresent),
		NTotal:   int(counts.NTotal),
		Votes:    votes,
	}, vote, nil
}
