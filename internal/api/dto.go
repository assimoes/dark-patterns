package api

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

// Pattern is a meso pattern the panel marked present on a review, with its code and name.
type Pattern struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// AdjReview is one review in a run's adjudication queue.
type AdjReview struct {
	ID       string    `json:"id"`
	GameID   string    `json:"gameId"`
	VotedUp  bool      `json:"votedUp"`
	Language string    `json:"language"`
	Body     string    `json:"body"`
	Present  []Pattern `json:"present"`
}

// decisionsRequest is the body of POST /api/reviews/{reviewId}/decisions: a map of pattern code to
// "present" | "absent". Keying by code (not row id) keeps the frontend free of database ids.
type decisionsRequest struct {
	Decisions map[string]string `json:"decisions"`
}

// ModelStat is the number of distinct reviews one LLM model annotated (scoped to a population).
type ModelStat struct {
	Model     string `json:"model"`
	Annotated int    `json:"annotated"`
}

// PopulationCoverage is one population's slice of a game: how many of the game's reviews fall in the
// population and how many are annotated. The dashboard drills game -> populations with these.
type PopulationCoverage struct {
	PopulationID int32  `json:"populationId"`
	Label        string `json:"label"`
	Reviews      int    `json:"reviews"`
	Annotated    int    `json:"annotated"`
}

// PanelMember is one annotator on a population's panel — an llm model or a human.
type PanelMember struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

// Population is one curated population as a pick-list row: its size and a human label (its description
// when set, otherwise a stable "Population #<id>").
type Population struct {
	ID          int    `json:"id"`
	Label       string `json:"label"`
	Modality    string `json:"modality"`
	Individuals int    `json:"individuals"`
	CreatedAt   string `json:"createdAt"`
}

// PopulationGame is one game's slice of a population: how many of the population's reviews belong to the
// game and how many are annotated, with the game's display name.
type PopulationGame struct {
	GameID    string `json:"gameId"`
	Name      string `json:"name"`
	Reviews   int    `json:"reviews"`
	Annotated int    `json:"annotated"`
}

// PopulationRun is one run that worked a population, as a compact reference for the population detail.
type PopulationRun struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	RunType string `json:"runType"`
}

// PopulationDetail is the body of GET /api/populations/{populationId}: the population header, its
// per-game breakdown, and the runs that worked it.
type PopulationDetail struct {
	ID        int              `json:"id"`
	Label     string           `json:"label"`
	Modality  string           `json:"modality"`
	CreatedAt string           `json:"createdAt"`
	PerGame   []PopulationGame `json:"perGame"`
	Runs      []PopulationRun  `json:"runs"`
}

// Annotator is one annotator as a form option: its kind, label, and the model name for llm annotators
// (null for humans).
type Annotator struct {
	ID    int     `json:"id"`
	Kind  string  `json:"kind"`
	Label string  `json:"label"`
	Model *string `json:"model"`
}

// Prompt is one prompt as a form option: its id, name, version and modality.
type Prompt struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Version  int    `json:"version"`
	Modality string `json:"modality"`
}

// RunSummary is one run in the run list: enough to recognise and pick it, including the panel size and a
// human label ("<run_type> #<id>").
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

// RunDetail is the body of GET /api/runs/{runId}: the run header, its frozen panel, and whether an
// adjudication sample has been drawn for it.
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

// GameSummary is one game with its presentation metadata and its review progress (reviews in scope and
// how many are annotated), the row a games list/grid renders.
type GameSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Short        string `json:"short"`
	Monetization string `json:"monetization"`
	Color        string `json:"color"`
	Reviews      int    `json:"reviews"`
	Annotated    int    `json:"annotated"`
}
