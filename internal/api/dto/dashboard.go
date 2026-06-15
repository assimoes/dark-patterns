package dto

import (
	"strconv"

	"github.com/assimoes/dsr/internal/db"
)

func NewGame(d db.GameDisplay) Game {
	return Game{
		ID:           GameID(d.ExternalGameID),
		Name:         d.Name,
		Short:        d.Short,
		Monetization: d.Monetization,
		Color:        d.DisplayColor,
	}
}

func NewPopulationStat(d db.CountIndividualsPerGameRow) PopulationStat {
	return PopulationStat{
		GameID:      GameID(d.ExternalGameID),
		Individuals: int(d.Individuals),
	}
}

func NewMember(d db.ListMembersForRunRow) Member {
	return Member{
		Kind:  d.Kind,
		Label: d.Label,
	}
}

func NewReviewStat(d db.ReviewStatsPerGameRow) ReviewStat {
	return ReviewStat{
		GameID:    GameID(d.ExternalGameID),
		RunID:     int(d.RunID),
		Prompt:    d.Prompt,
		Reviews:   int(d.Reviews),
		Annotated: int(d.Annotated),
	}
}

func NewRun(d db.ListRunsForDashboardRow, members []Member) Run {
	return Run{
		ID:         strconv.FormatInt(int64(d.ID), 10),
		Label:      RunLabel(d.RunType, d.ID),
		Population: d.Population,
		CreatedAt:  RFC3339(d.CreatedAt),
		Members:    members,
	}
}

// Game is one curated Steam game with its presentation metadata.
type Game struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Short        string `json:"short"`
	Monetization string `json:"monetization"`
	Color        string `json:"color"`
}

// PopulationStat is the number of curated individuals (reviews in scope) for a game.
type PopulationStat struct {
	GameID      string `json:"gameId"`
	Individuals int    `json:"individuals"`
}

// Member is one panel member of a run: an llm model or a human annotator.
type Member struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

// Run is an annotation run with the panel that produced it.
type Run struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Population string   `json:"population"`
	CreatedAt  string   `json:"createdAt"`
	Members    []Member `json:"members"`
}

// ReviewStat is per-game progress: reviews in scope and how many are annotated.
type ReviewStat struct {
	GameID    string `json:"gameId"`
	RunID     int    `json:"runId"`
	Prompt    string `json:"prompt"`
	Reviews   int    `json:"reviews"`
	Annotated int    `json:"annotated"`
}

// DashboardData is the single payload behind GET /api/dashboard.
type DashboardData struct {
	Games       []Game           `json:"games"`
	Populations []PopulationStat `json:"populations"`
	Runs        []Run            `json:"runs"`
	ReviewStats []ReviewStat     `json:"reviewStats"`
}
