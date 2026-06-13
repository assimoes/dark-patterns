package dto

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
