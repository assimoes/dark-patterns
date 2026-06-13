package dto

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
