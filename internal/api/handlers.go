package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/assimoes/dsr/internal/adjudicate"
	"github.com/assimoes/dsr/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// dashboard returns the single summary payload behind the landing page: the curated games, the
// per-game population sizes, the runs with their panels, and per-game annotation progress. Each
// section is its own focused query; the handler stitches them into the response shape.
func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	displays, err := s.q.ListGameDisplays(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load games", err)
		return
	}

	games := make([]Game, 0, len(displays))
	for _, d := range displays {
		games = append(games, Game{
			ID:           gameID(d.ExternalGameID),
			Name:         d.Name,
			Short:        d.Short,
			Monetization: d.Monetization,
			Color:        d.DisplayColor,
		})
	}

	pops, err := s.q.CountIndividualsPerGame(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load populations", err)
		return
	}

	populations := make([]PopulationStat, 0, len(pops))
	for _, p := range pops {
		populations = append(populations, PopulationStat{
			GameID:      gameID(p.ExternalGameID),
			Individuals: int(p.Individuals),
		})
	}

	runRows, err := s.q.ListRunsForDashboard(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load runs", err)
		return
	}

	runs := make([]Run, 0, len(runRows))
	for _, rr := range runRows {
		members, err := s.q.ListMembersForRun(ctx, rr.ID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "load run members", err)
			return
		}

		out := make([]Member, 0, len(members))
		for _, m := range members {
			out = append(out, Member{Kind: m.Kind, Label: m.Label})
		}

		runs = append(runs, Run{
			ID:         strconv.FormatInt(int64(rr.ID), 10),
			Label:      runLabel(rr.RunType, rr.ID),
			Population: rr.Population,
			CreatedAt:  rfc3339(rr.CreatedAt),
			Members:    out,
		})
	}

	statRows, err := s.q.ReviewStatsPerGame(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load review stats", err)
		return
	}

	reviewStats := make([]ReviewStat, 0, len(statRows))
	for _, st := range statRows {
		reviewStats = append(reviewStats, ReviewStat{
			GameID:    gameID(st.ExternalGameID),
			RunID:     int(st.RunID),
			Prompt:    st.Prompt,
			Reviews:   int(st.Reviews),
			Annotated: int(st.Annotated),
		})
	}

	s.writeJSON(w, http.StatusOK, DashboardData{
		Games:       games,
		Populations: populations,
		Runs:        runs,
		ReviewStats: reviewStats,
	})
}

// gameModels returns per-LLM-model completed annotation counts for one game. The {gameId} path
// value is the external_game_id the dashboard handed out.
func (s *Server) gameModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseGameID(r.PathValue("gameId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid gameId", err)
		return
	}

	rows, err := s.q.ModelStatsForGame(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load model stats", err)
		return
	}

	out := make([]ModelStat, 0, len(rows))
	for _, m := range rows {
		out = append(out, ModelStat{Model: m.Model, Annotated: int(m.Annotated)})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// runReviews returns the adjudication queue for a run: one entry per annotated review, with the
// patterns the panel marked present. The present set is computed per review by a majority query.
func (s *Server) runReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	runID, err := parseInt32(r.PathValue("runId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid runId", err)
		return
	}

	rows, err := s.q.ListRunReviews(ctx, runID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load run reviews", err)
		return
	}

	out := make([]AdjReview, 0, len(rows))
	for _, rv := range rows {
		present, err := s.q.ListPresentPatternsForReview(ctx, db.ListPresentPatternsForReviewParams{
			RunID:        runID,
			IndividualID: rv.IndividualID,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "load present patterns", err)
			return
		}

		patterns := make([]Pattern, 0, len(present))
		for _, p := range present {
			patterns = append(patterns, Pattern{Code: p.Code, Name: p.Name})
		}

		out = append(out, AdjReview{
			ID:       strconv.FormatInt(rv.IndividualID, 10),
			GameID:   gameID(rv.ExternalGameID),
			VotedUp:  rv.VotedUp,
			Language: rv.Lang,
			Body:     rv.Body,
			Present:  patterns,
		})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// reviewDecisions records the auditor's present/absent calls for one review. The {reviewId} path
// value is the individual id. Each decision is written via UpsertAdjudication with the panel state
// frozen at decision time, exactly as the adjudication pipeline does it; the whole set lands in one
// transaction so a partial save never happens. Responds 204 on success.
func (s *Server) reviewDecisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	individualID, err := strconv.ParseInt(r.PathValue("reviewId"), 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid reviewId", err)
		return
	}

	var req decisionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad json body", err)
		return
	}

	// Resolve the panel run the auditor is actually working under — the ?run= the screen sends — and
	// derive the gold run and taxonomy version from IT, not from "newest panel run". A population can
	// hold several panel runs of different taxonomy versions; resolving independently here would write
	// the decisions against a different run's gold run and pattern-version ids than the read maps with,
	// so the saved labels would never show up on revisit.
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

// buildSeed reconstructs the panel verdict for one (panel run, review, pattern) from a fresh trusted
// read and returns both the frozen seed and the panel's majority vote, so a decision can be recorded
// as a confirmation or a replacement of what the panel said.
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

// gamePopulations returns the populations that contain a game's reviews, each with the game's slice
// (reviews in scope and how many are annotated). The drilldown groups by population so the counts
// never sum across populations the way a flat per-game total does.
func (s *Server) gamePopulations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseGameID(r.PathValue("gameId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid gameId", err)
		return
	}

	rows, err := s.q.ListPopulationsForGame(ctx, id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load populations", err)
		return
	}

	out := make([]PopulationCoverage, 0, len(rows))
	for _, p := range rows {
		out = append(out, PopulationCoverage{
			PopulationID: p.PopulationID,
			Label:        populationLabel(p.PopulationID, p.Description),
			Reviews:      int(p.Reviews),
			Annotated:    int(p.Annotated),
		})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// gamePopulationModels returns per-LLM-model distinct-review counts for one game WITHIN one
// population, so the models are directly comparable. Both ids come from the path.
func (s *Server) gamePopulationModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gid, err := parseGameID(r.PathValue("gameId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid gameId", err)
		return
	}

	pid, err := parseInt32(r.PathValue("populationId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid populationId", err)
		return
	}

	rows, err := s.q.ModelStatsForGamePopulation(ctx, db.ModelStatsForGamePopulationParams{
		ExternalGameID: gid,
		PopulationID:   pid,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load model stats", err)
		return
	}

	out := make([]ModelStat, 0, len(rows))
	for _, m := range rows {
		out = append(out, ModelStat{Model: m.Model, Annotated: int(m.Annotated)})
	}

	s.writeJSON(w, http.StatusOK, out)
}

func (s *Server) gameRunModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gid, err := parseGameID(r.PathValue("gameId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid gameId", err)
		return
	}

	rid, err := parseInt32(r.PathValue("runId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid runId", err)
		return
	}

	rows, err := s.q.ModelStatsForGameRun(ctx, db.ModelStatsForGameRunParams{
		ExternalGameID: gid,
		RunID:          rid,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load model stats", err)
		return
	}

	out := make([]ModelStat, 0, len(rows))
	for _, m := range rows {
		out = append(out, ModelStat{Model: m.Model, Annotated: int(m.Annotated)})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// populationPanel returns the annotators frozen onto a population's runs — the panel that worked it.
func (s *Server) populationPanel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pid, err := parseInt32(r.PathValue("populationId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid populationId", err)
		return
	}

	rows, err := s.q.PanelForPopulation(ctx, pid)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load panel", err)
		return
	}

	out := make([]PanelMember, 0, len(rows))
	for _, m := range rows {
		out = append(out, PanelMember{Kind: m.Kind, Label: m.Label})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// populationLabel is the human label for a population row: its description when set, otherwise a
// stable "Population #<id>".
func populationLabel(id int32, desc *string) string {
	if desc != nil && *desc != "" {
		return *desc
	}

	return "Population #" + strconv.FormatInt(int64(id), 10)
}

// gameID renders an external_game_id as the string id the frontend uses.
func gameID(id int32) string {
	return strconv.FormatInt(int64(id), 10)
}

// parseGameID turns the {gameId} path value back into an external_game_id.
func parseGameID(s string) (int32, error) {
	return parseInt32(s)
}

// parseInt32 parses a base-10 int32 from a path value.
func parseInt32(s string) (int32, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(n), nil
}

// parseLabel maps the frontend's "present"/"absent" string to a boolean final label.
func parseLabel(s string) (label, ok bool) {
	switch s {
	case "present":
		return true, true
	case "absent":
		return false, true
	default:
		return false, false
	}
}

// runLabel is the human label the frontend shows for a run. The runs table has no name column, so we
// derive a stable label from the run type and id.
func runLabel(runType string, id int32) string {
	return runType + " #" + strconv.FormatInt(int64(id), 10)
}

// rfc3339 renders a nullable timestamp as an RFC3339 string, or empty when the column is null.
func rfc3339(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}

	return ts.Time.UTC().Format(time.RFC3339)
}
