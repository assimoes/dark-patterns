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
