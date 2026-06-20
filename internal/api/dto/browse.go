package dto

import "github.com/assimoes/dsr/internal/db"

func NewAnnotator(a db.ListAnnotatorsWithModelRow) Annotator {
	return Annotator{
		ID:       int(a.ID),
		Kind:     a.Kind,
		Label:    a.Label,
		Model:    a.Model,
		RefCount: int(a.RefCount),
	}
}

func NewPrompt(a db.ListPromptsRow) Prompt {
	return Prompt{
		ID:       int(a.ID),
		Name:     a.Name,
		Version:  int(a.Version),
		Modality: a.Modality,
		RunCount: int(a.RunCount),
	}
}

func NewPopulation(p db.ListPopulationsRow) Population {
	return Population{
		ID:          int(p.ID),
		Label:       PopulationLabel(p.ID, p.Description),
		Modality:    p.Modality,
		Individuals: int(p.Individuals),
		CreatedAt:   RFC3339(p.CreatedAt),
	}
}

func NewPopulationGame(g db.PopulationPerGameRow, name string) PopulationGame {
	return PopulationGame{
		GameID:    GameID(g.ExternalGameID),
		Name:      name,
		Reviews:   int(g.Reviews),
		Annotated: int(g.Annotated),
	}
}

func NewPopulationRun(r db.Run) PopulationRun {
	return PopulationRun{
		ID:      int(r.ID),
		Label:   RunLabel(r.RunType, r.ID),
		RunType: r.RunType,
	}
}

func NewPopulationDetail(p db.GetPopulationRow, perGame []PopulationGame, runs []PopulationRun) PopulationDetail {
	return PopulationDetail{
		ID:        int(p.ID),
		Label:     PopulationLabel(p.ID, p.Description),
		Modality:  p.Modality,
		CreatedAt: RFC3339(p.CreatedAt),
		PerGame:   perGame,
		Runs:      runs,
	}
}

func NewRunSummary(r db.ListRunsRow) RunSummary {
	return RunSummary{
		ID:              int(r.ID),
		RunType:         r.RunType,
		Label:           RunLabel(r.RunType, r.ID),
		Population:      r.Population,
		PopulationID:    int(r.PopulationID),
		PromptID:        int(r.PromptID),
		TaxonomyVersion: IntPtr(r.TaxonomyVersion),
		CreatedAt:       RFC3339(r.CreatedAt),
		PanelSize:       int(r.PanelSize),
	}
}

func NewRunDetail(run db.Run, population string, panel []Member, hasSample bool) RunDetail {
	return RunDetail{
		ID:              int(run.ID),
		RunType:         run.RunType,
		Population:      population,
		PopulationID:    int(run.PopulationID),
		PromptID:        int(run.PromptID),
		TaxonomyVersion: IntPtr(run.TaxonomyVersion),
		CreatedAt:       RFC3339(run.CreatedAt),
		Panel:           panel,
		HasSample:       hasSample,
	}
}

func NewGameSummary(d db.GameDisplay, reviews, annotated, artifacts int, descriptionStatus string) GameSummary {
	return GameSummary{
		ID:                GameID(d.ExternalGameID),
		Name:              d.Name,
		Short:             d.Short,
		Monetization:      d.Monetization,
		Color:             d.DisplayColor,
		Reviews:           reviews,
		Annotated:         annotated,
		Artifacts:         artifacts,
		SourceRefs:        DecodeRefs(d.SourceRefs),
		DescriptionStatus: descriptionStatus,
	}
}

// Population is one curated population as a pick-list row.
type Population struct {
	ID          int    `json:"id"`
	Label       string `json:"label"`
	Modality    string `json:"modality"`
	Individuals int    `json:"individuals"`
	CreatedAt   string `json:"createdAt"`
}

// PopulationGame is one games slice of a population: reviews in scope, how many annotated, plus name.
type PopulationGame struct {
	GameID    string `json:"gameId"`
	Name      string `json:"name"`
	Reviews   int    `json:"reviews"`
	Annotated int    `json:"annotated"`
}

// PopulationRun is a compact reference to one run that worked a population.
type PopulationRun struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	RunType string `json:"runType"`
}

// PopulationDetail is the body of GET /api/populations/{populationId}: header, per-game breakdown, runs.
type PopulationDetail struct {
	ID        int              `json:"id"`
	Label     string           `json:"label"`
	Modality  string           `json:"modality"`
	CreatedAt string           `json:"createdAt"`
	PerGame   []PopulationGame `json:"perGame"`
	Runs      []PopulationRun  `json:"runs"`
}

// Annotator is one annotator as a form option (model name for llm, null for humans). refCount is how
// many runs/annotations reference it; nonzero means it is frozen and cannot be deleted.
type Annotator struct {
	ID       int     `json:"id"`
	Kind     string  `json:"kind"`
	Label    string  `json:"label"`
	Model    *string `json:"model"`
	RefCount int     `json:"refCount"`
}

// Prompt is one prompt as a form option. runCount is how many runs use it; nonzero means it is frozen
// and cannot be edited or deleted.
type Prompt struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Version  int    `json:"version"`
	Modality string `json:"modality"`
	RunCount int    `json:"runCount"`
}

// RunSummary is one run in the run list: enough to recognise and pick it.
type RunSummary struct {
	ID              int    `json:"id"`
	RunType         string `json:"runType"`
	Label           string `json:"label"`
	Population      string `json:"population"`
	PopulationID    int    `json:"populationId"`
	PromptID        int    `json:"promptId"`
	TaxonomyVersion *int   `json:"taxonomyVersion"`
	CreatedAt       string `json:"createdAt"`
	PanelSize       int    `json:"panelSize"`
}

// RunDetail is the body of GET /api/runs/{runId}: header, frozen panel, and whether a sample was drawn.
type RunDetail struct {
	ID              int      `json:"id"`
	RunType         string   `json:"runType"`
	Population      string   `json:"population"`
	PopulationID    int      `json:"populationId"`
	PromptID        int      `json:"promptId"`
	TaxonomyVersion *int     `json:"taxonomyVersion"`
	CreatedAt       string   `json:"createdAt"`
	Panel           []Member `json:"panel"`
	HasSample       bool     `json:"hasSample"`
}

// GameSummary is one game with its display metadata and review progress, the row a games grid renders.
type GameSummary struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Short             string            `json:"short"`
	Monetization      string            `json:"monetization"`
	Color             string            `json:"color"`
	Reviews           int               `json:"reviews"`
	Annotated         int               `json:"annotated"`
	Artifacts         int               `json:"artifacts"`
	SourceRefs        map[string]string `json:"sourceRefs"`
	DescriptionStatus string            `json:"descriptionStatus"`
}
