package dto

import "github.com/assimoes/dsr/internal/db"

func NewPanelMember(d db.PanelForPopulationRow) PanelMember {
	return PanelMember{
		Kind:  d.Kind,
		Label: d.Label,
	}
}

func NewModelStat(model string, annotated int32) ModelStat {
	return ModelStat{
		Model:     model,
		Annotated: int(annotated),
	}
}

func NewPopulationCoverage(p db.ListPopulationsForGameRow) PopulationCoverage {
	return PopulationCoverage{
		PopulationID: p.PopulationID,
		Label:        PopulationLabel(p.PopulationID, p.Description),
		Reviews:      int(p.Reviews),
		Annotated:    int(p.Annotated),
	}
}

// ModelStat is the number of distinct reviews one LLM model annotated (scoped to a population).
type ModelStat struct {
	Model     string `json:"model"`
	Annotated int    `json:"annotated"`
}

// PopulationCoverage is one populations slice of a game: reviews in scope and how many annotated.
type PopulationCoverage struct {
	PopulationID int32  `json:"populationId"`
	Label        string `json:"label"`
	Reviews      int    `json:"reviews"`
	Annotated    int    `json:"annotated"`
}

// PanelMember is one annotator on a populations panel, an llm model or a human.
type PanelMember struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}
