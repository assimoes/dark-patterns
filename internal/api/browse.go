package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/jackc/pgx/v5"
)

// listPopulations returns every population as a pick-list row. GET /api/populations.
func (s *Server) listPopulations(w http.ResponseWriter, r *http.Request) {
	writeList(s, w, r, s.q.ListPopulations, "load populations", dto.NewPopulation)
}

// populationDetail returns one populations header, per-game breakdown, and the runs that worked it.
// GET /api/populations/{populationId}.
func (s *Server) populationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pid, err := parseInt32(r.PathValue("populationId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid populationId", err)
		return
	}

	pop, err := s.q.GetPopulation(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.writeError(w, http.StatusNotFound, "population not found", err)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "load population", err)
		return
	}

	names, err := s.gameNames(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load games", err)
		return
	}

	perGameRows, err := s.q.PopulationPerGame(ctx, pid)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load per-game stats", err)
		return
	}

	perGame := make([]dto.PopulationGame, 0, len(perGameRows))
	for _, g := range perGameRows {
		perGame = append(perGame, dto.NewPopulationGame(g, names[g.ExternalGameID]))
	}

	runRows, err := s.q.ListRunsByPopulation(ctx, pid)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load population runs", err)
		return
	}

	runs := make([]dto.PopulationRun, 0, len(runRows))
	for _, rr := range runRows {
		runs = append(runs, dto.NewPopulationRun(rr))
	}

	s.writeJSON(w, http.StatusOK, dto.NewPopulationDetail(pop, perGame, runs))
}

// listAnnotators returns every annotator as a form option (model name for llm, null for humans).
// GET /api/annotators.
func (s *Server) listAnnotators(w http.ResponseWriter, r *http.Request) {
	writeList(s, w, r, s.q.ListAnnotatorsWithModel, "load annotators", dto.NewAnnotator)
}

// listPrompts returns every prompt as a form option. GET /api/prompts.
func (s *Server) listPrompts(w http.ResponseWriter, r *http.Request) {
	writeList(s, w, r, s.q.ListPrompts, "load prompts", dto.NewPrompt)
}

// listRuns returns every run with enough to recognise and pick it. GET /api/runs.
func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	writeList(s, w, r, s.q.ListRuns, "load runs", dto.NewRunSummary)
}

// runDetail returns one runs header, its frozen panel, and whether a sample was drawn. GET /api/runs/{runId}.
func (s *Server) runDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	runID, err := parseInt32(r.PathValue("runId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid runId", err)
		return
	}

	run, err := s.q.GetRun(ctx, runID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.writeError(w, http.StatusNotFound, "run not found", err)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "load run", err)
		return
	}

	pop, err := s.q.GetPopulation(ctx, run.PopulationID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load run population", err)
		return
	}

	members, err := s.q.ListMembersForRun(ctx, runID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load run panel", err)
		return
	}

	panel := make([]dto.Member, 0, len(members))
	for _, m := range members {
		panel = append(panel, dto.Member{Kind: m.Kind, Label: m.Label})
	}

	hasSample, err := s.q.RunHasSample(ctx, runID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load run sample status", err)
		return
	}

	s.writeJSON(w, http.StatusOK, dto.NewRunDetail(run, pop.Modality, panel, hasSample))
}

// listGames returns every curated game with its display metadata and review progress, joined by
// external_game_id. GET /api/games.
func (s *Server) listGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	displays, err := s.q.ListGameDisplays(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load games", err)
		return
	}

	statRows, err := s.q.GameReviewTotals(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load review stats", err)
		return
	}

	artifactRows, err := s.q.GameArtifactTotals(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load artifact totals", err)
		return
	}

	type stat struct{ reviews, annotated int }
	stats := make(map[int32]stat, len(statRows))
	for _, st := range statRows {
		stats[st.ExternalGameID] = stat{reviews: int(st.Reviews), annotated: int(st.Annotated)}
	}

	artifacts := make(map[int32]int, len(artifactRows))
	for _, a := range artifactRows {
		artifacts[a.ExternalGameID] = int(a.Artifacts)
	}

	stateRows, err := s.q.ListGameDescriptionStates(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load description states", err)
		return
	}

	descStatus := make(map[int32]string, len(stateRows))
	for _, st := range stateRows {
		descStatus[st.ExternalGameID] = st.Status
	}

	out := make([]dto.GameSummary, 0, len(displays))
	for _, d := range displays {
		st := stats[d.ExternalGameID]
		status := descStatus[d.ExternalGameID]
		if status == "" {
			status = "none"
		}
		out = append(out, dto.NewGameSummary(d, st.reviews, st.annotated, artifacts[d.ExternalGameID], status))
	}

	s.writeJSON(w, http.StatusOK, out)
}

// gameNames builds an external_game_id -> display name lookup, so the per-game query needs no join.
func (s *Server) gameNames(ctx context.Context) (map[int32]string, error) {
	displays, err := s.q.ListGameDisplays(ctx)
	if err != nil {
		return nil, err
	}

	names := make(map[int32]string, len(displays))
	for _, d := range displays {
		names[d.ExternalGameID] = d.Name
	}

	return names, nil
}
