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
	rows, err := s.q.ListPopulations(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load populations", err)
		return
	}

	out := make([]dto.Population, 0, len(rows))
	for _, p := range rows {
		out = append(out, dto.Population{
			ID:          int(p.ID),
			Label:       populationLabel(p.ID, p.Description),
			Modality:    p.Modality,
			Individuals: int(p.Individuals),
			CreatedAt:   rfc3339(p.CreatedAt),
		})
	}

	s.writeJSON(w, http.StatusOK, out)
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
		perGame = append(perGame, dto.PopulationGame{
			GameID:    gameID(g.ExternalGameID),
			Name:      names[g.ExternalGameID],
			Reviews:   int(g.Reviews),
			Annotated: int(g.Annotated),
		})
	}

	runRows, err := s.q.ListRunsByPopulation(ctx, pid)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load population runs", err)
		return
	}

	runs := make([]dto.PopulationRun, 0, len(runRows))
	for _, rr := range runRows {
		runs = append(runs, dto.PopulationRun{
			ID:      int(rr.ID),
			Label:   runLabel(rr.RunType, rr.ID),
			RunType: rr.RunType,
		})
	}

	s.writeJSON(w, http.StatusOK, dto.PopulationDetail{
		ID:        int(pop.ID),
		Label:     populationLabel(pop.ID, pop.Description),
		Modality:  pop.Modality,
		CreatedAt: rfc3339(pop.CreatedAt),
		PerGame:   perGame,
		Runs:      runs,
	})
}

// listAnnotators returns every annotator as a form option (model name for llm, null for humans).
// GET /api/annotators.
func (s *Server) listAnnotators(w http.ResponseWriter, r *http.Request) {
	rows, err := s.q.ListAnnotatorsWithModel(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load annotators", err)
		return
	}

	out := make([]dto.Annotator, 0, len(rows))
	for _, a := range rows {
		out = append(out, dto.Annotator{
			ID:    int(a.ID),
			Kind:  a.Kind,
			Label: a.Label,
			Model: a.Model,
		})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// listPrompts returns every prompt as a form option. GET /api/prompts.
func (s *Server) listPrompts(w http.ResponseWriter, r *http.Request) {
	rows, err := s.q.ListPrompts(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load prompts", err)
		return
	}

	out := make([]dto.Prompt, 0, len(rows))
	for _, p := range rows {
		out = append(out, dto.Prompt{
			ID:       int(p.ID),
			Name:     p.Name,
			Version:  int(p.Version),
			Modality: p.Modality,
		})
	}

	s.writeJSON(w, http.StatusOK, out)
}

// listRuns returns every run with enough to recognise and pick it. GET /api/runs.
func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	rows, err := s.q.ListRuns(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load runs", err)
		return
	}

	out := make([]dto.RunSummary, 0, len(rows))
	for _, rr := range rows {
		out = append(out, dto.RunSummary{
			ID:              int(rr.ID),
			RunType:         rr.RunType,
			Label:           runLabel(rr.RunType, rr.ID),
			Population:      rr.Population,
			PopulationID:    int(rr.PopulationID),
			PromptID:        int(rr.PromptID),
			TaxonomyVersion: intPtr(rr.TaxonomyVersion),
			CreatedAt:       rfc3339(rr.CreatedAt),
			PanelSize:       int(rr.PanelSize),
		})
	}

	s.writeJSON(w, http.StatusOK, out)
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

	// no sample is the expected no-rows case, not a failure; swallow ErrNoRows into false
	hasSample := false
	// if _, err := s.q.GetLatestSampleForRun(ctx, runID); err != nil {
	// 	if !errors.Is(err, pgx.ErrNoRows) {
	// 		s.writeError(w, http.StatusInternalServerError, "load sample", err)
	// 		return
	// 	}
	// 	hasSample = false
	// }

	s.writeJSON(w, http.StatusOK, dto.RunDetail{
		ID:              int(run.ID),
		RunType:         run.RunType,
		Population:      pop.Modality,
		PopulationID:    int(run.PopulationID),
		PromptID:        int(run.PromptID),
		TaxonomyVersion: intPtr(run.TaxonomyVersion),
		CreatedAt:       rfc3339(run.CreatedAt),
		Panel:           panel,
		HasSample:       hasSample,
	})
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

	type stat struct{ reviews, annotated int }
	stats := make(map[int32]stat, len(statRows))
	for _, st := range statRows {
		stats[st.ExternalGameID] = stat{reviews: int(st.Reviews), annotated: int(st.Annotated)}
	}

	out := make([]dto.GameSummary, 0, len(displays))
	for _, d := range displays {
		st := stats[d.ExternalGameID]
		out = append(out, dto.GameSummary{
			ID:           gameID(d.ExternalGameID),
			Name:         d.Name,
			Short:        d.Short,
			Monetization: d.Monetization,
			Color:        d.DisplayColor,
			Reviews:      st.reviews,
			Annotated:    st.annotated,
		})
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

// intPtr maps a nullable *int32 to the *int the DTOs expose, preserving null.
func intPtr(v *int32) *int {
	if v == nil {
		return nil
	}

	n := int(*v)
	return &n
}
