package api

import (
	"context"
	"net/http"

	"github.com/assimoes/dsr/internal/api/dto"
	"github.com/assimoes/dsr/internal/db"
)

// gamePopulations returns the populations holding a games reviews, each with the games slice.
// grouping by population keeps counts from summing across populations like a flat total would.
func (s *Server) gamePopulations(w http.ResponseWriter, r *http.Request) {
	id, err := parseGameID(r.PathValue("gameId"))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid gameId", err)
		return
	}
	writeList(s, w, r, func(ctx context.Context) ([]db.ListPopulationsForGameRow, error) {
		return s.q.ListPopulationsForGame(ctx, id)
	}, "load populations", dto.NewPopulationCoverage)
}

// gamePopulationModels returns per-model distinct-review counts for one game within one population,
// so the models are comparable. Both ids come from the path.
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

	out := make([]dto.ModelStat, 0, len(rows))
	for _, m := range rows {
		out = append(out, dto.NewModelStat(m.Model, m.Annotated))
	}

	s.writeJSON(w, http.StatusOK, out)
}

// gameRunModels is gamePopulationModels but scoped to one run instead of a whole population, so you see
// what each model did on this exact run. both ids come from the path.
func (s *Server) gameRunModels(w http.ResponseWriter, r *http.Request) {

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

	writeList(s, w, r, func(ctx context.Context) ([]db.ModelStatsForGameRunRow, error) {
		return s.q.ModelStatsForGameRun(ctx, db.ModelStatsForGameRunParams{
			ExternalGameID: gid,
			RunID:          rid,
		})
	}, "load model stats", func(m db.ModelStatsForGameRunRow) dto.ModelStat {
		return dto.NewModelStat(m.Model, m.Annotated)
	})
}

// populationPanel returns the annotators frozen onto a populations runs.
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

	out := make([]dto.PanelMember, 0, len(rows))
	for _, m := range rows {
		out = append(out, dto.NewPanelMember(m))
	}

	s.writeJSON(w, http.StatusOK, out)
}
