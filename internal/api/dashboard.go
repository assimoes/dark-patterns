package api

import (
	"net/http"
	"strconv"

	"github.com/assimoes/dsr/internal/api/dto"
)

// dashboard serves GET /api/dashboard: games, population sizes, runs, and review progress.
func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	displays, err := s.q.ListGameDisplays(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load games", err)
		return
	}

	games := make([]dto.Game, 0, len(displays))
	for _, d := range displays {
		games = append(games, dto.Game{
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

	populations := make([]dto.PopulationStat, 0, len(pops))
	for _, p := range pops {
		populations = append(populations, dto.PopulationStat{
			GameID:      gameID(p.ExternalGameID),
			Individuals: int(p.Individuals),
		})
	}

	runRows, err := s.q.ListRunsForDashboard(ctx)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "load runs", err)
		return
	}

	runs := make([]dto.Run, 0, len(runRows))
	for _, rr := range runRows {
		members, err := s.q.ListMembersForRun(ctx, rr.ID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "load run members", err)
			return
		}

		out := make([]dto.Member, 0, len(members))
		for _, m := range members {
			out = append(out, dto.Member{Kind: m.Kind, Label: m.Label})
		}

		runs = append(runs, dto.Run{
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

	reviewStats := make([]dto.ReviewStat, 0, len(statRows))
	for _, st := range statRows {
		reviewStats = append(reviewStats, dto.ReviewStat{
			GameID:    gameID(st.ExternalGameID),
			RunID:     int(st.RunID),
			Prompt:    st.Prompt,
			Reviews:   int(st.Reviews),
			Annotated: int(st.Annotated),
		})
	}

	s.writeJSON(w, http.StatusOK, dto.DashboardData{
		Games:       games,
		Populations: populations,
		Runs:        runs,
		ReviewStats: reviewStats,
	})
}
