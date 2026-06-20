package dto

import (
	"strconv"

	"github.com/assimoes/dsr/internal/db"
)

// CreateComparisonRequest is the body of POST /api/comparisons.
type CreateComparisonRequest struct {
	Label  string `json:"label"`
	RunAID int    `json:"runAId"`
	RunBID int    `json:"runBId"`
}

// Comparison is a saved run-vs-run comparison as a list row.
type Comparison struct {
	ID        int    `json:"id"`
	Label     string `json:"label"`
	RunAID    int    `json:"runAId"`
	RunBID    int    `json:"runBId"`
	CreatedAt string `json:"createdAt"`
}

func NewComparison(c db.Comparison) Comparison {
	return Comparison{
		ID:        int(c.ID),
		Label:     c.Label,
		RunAID:    int(c.RunAID),
		RunBID:    int(c.RunBID),
		CreatedAt: RFC3339(c.CreatedAt),
	}
}

// AttributeDiff is one run attribute side by side and whether the two runs match on it — informational
// context for reading the divergence, not a constraint.
type AttributeDiff struct {
	Attribute string `json:"attribute"`
	A         string `json:"a"`
	B         string `json:"b"`
	Equal     bool   `json:"equal"`
}

// RunRef identifies a run in the comparison header.
type RunRef struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	RunType string `json:"runType"`
}

// ModelAgreement is how often one panel model's flagged patterns matched across the two runs.
type ModelAgreement struct {
	Model         string `json:"model"`
	AgreedPresent int    `json:"agreedPresent"`
	Flips         int    `json:"flips"`
}

// ComparisonDetail backs the comparison show page: the runs, an informational attribute diff, the
// per-model agreement, and the review totals.
type ComparisonDetail struct {
	Comparison
	RunA            RunRef           `json:"runA"`
	RunB            RunRef           `json:"runB"`
	Attributes      []AttributeDiff  `json:"attributes"`
	ModelAgreement  []ModelAgreement `json:"modelAgreement"`
	ReviewsTotal    int              `json:"reviewsTotal"`
	ReviewsDiverged int              `json:"reviewsDiverged"`
}

// ComparisonReviewRow is one shared review in the divergence worklist. status is diverged or converged.
type ComparisonReviewRow struct {
	ReviewID string `json:"reviewId"`
	GameID   string `json:"gameId"`
	Flips    int    `json:"flips"`
	Status   string `json:"status"`
}

func NewComparisonReviewRow(d db.ComparisonReviewDivergenceRow) ComparisonReviewRow {
	status := "converged"
	if d.Flips > 0 {
		status = "diverged"
	}
	return ComparisonReviewRow{
		ReviewID: strconv.FormatInt(d.IndividualID, 10),
		GameID:   GameID(d.ExternalGameID),
		Flips:    int(d.Flips),
		Status:   status,
	}
}
