package dto

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

// Annotator is one annotator as a form option (model name for llm, null for humans).
type Annotator struct {
	ID    int     `json:"id"`
	Kind  string  `json:"kind"`
	Label string  `json:"label"`
	Model *string `json:"model"`
}

// Prompt is one prompt as a form option.
type Prompt struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Version  int    `json:"version"`
	Modality string `json:"modality"`
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
	ID           string `json:"id"`
	Name         string `json:"name"`
	Short        string `json:"short"`
	Monetization string `json:"monetization"`
	Color        string `json:"color"`
	Reviews      int    `json:"reviews"`
	Annotated    int    `json:"annotated"`
}
