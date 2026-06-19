package dto

import "github.com/assimoes/dsr/internal/db"

// DecisionsRequest is the body of POST /api/reviews/{reviewId}/decisions: pattern code -> "present" |
// "absent". Keying by code, not row id, keeps the frontend free of database ids.
type DecisionsRequest struct {
	Decisions map[string]string `json:"decisions"`
}

// CreateSampleRequest is the body of POST /api/adjudication-samples.
type CreateSampleRequest struct {
	PanelRunID      int32  `json:"panelRunId"`
	FlaggedMajority int    `json:"flaggedMajority"`
	FlaggedSplit    int    `json:"flaggedSplit"`
	Silent          int    `json:"silent"`
	Seed            *int64 `json:"seed"`
}

// StratumCounts is how many reviews were drawn into each stratum (min(target, stratum size)).
type StratumCounts struct {
	FlaggedMajority int `json:"flagged_majority"`
	FlaggedSplit    int `json:"flagged_split"`
	Silent          int `json:"silent"`
}

// CreateSampleResponse is the 201 body of POST /api/adjudication-samples.
type CreateSampleResponse struct {
	SampleID  int64         `json:"sampleId"`
	GoldRunID int32         `json:"goldRunId"`
	Seed      int64         `json:"seed"`
	Counts    StratumCounts `json:"counts"`
}

// SampleReview is one review in a persisted samples queue, with its stratum and decided-pattern count.
// modality tells the worklist whether it is paging text or image reviews.
type SampleReview struct {
	ID       string `json:"id"`
	GameID   string `json:"gameId"`
	Stratum  string `json:"stratum"`
	Modality string `json:"modality"`
	VotedUp  bool   `json:"votedUp"`
	Language string `json:"language"`
	Decided  int    `json:"decided"`
}

// SampleResponse is the body of GET /api/runs/{runId}/adjudication-sample.
type SampleResponse struct {
	SampleID  int64          `json:"sampleId"`
	GoldRunID int32          `json:"goldRunId"`
	Reviews   []SampleReview `json:"reviews"`
}

func NewDetection(patternCode string, v db.ListReviewDetectionsRow) Detection {
	return Detection{
		PatternCode: patternCode,
		Model:       v.ModelSlug,
		Evidence:    v.Evidence,
		Explanation: v.Explanation,
	}
}

// Detection is one panel models positive call on a pattern, with its own evidence and explanation.
type Detection struct {
	PatternCode string `json:"patternCode"`
	Model       string `json:"model"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
}

// AdjudicationReview is the full review an auditor sees. modality says how to render it.
type AdjudicationReview struct {
	ID          string          `json:"id"`
	GameID      string          `json:"gameId"`
	Modality    string          `json:"modality"`
	VotedUp     bool            `json:"votedUp"`
	Language    string          `json:"language"`
	Body        string          `json:"body"`
	ImageURI    string          `json:"imageUri"`
	Description string          `json:"description"`
	PanelModels []string        `json:"panelModels"`
	Detections  []Detection     `json:"detections"`
	GoldLabels  map[string]bool `json:"goldLabels"`
}

// BlindReview is the panel-free projection for a blind pass. it hides every panel vote, but it still
// carries the auditors own saved labels so a reopened review comes back with its decisions in place.
type BlindReview struct {
	ID         string          `json:"id"`
	GameID     string          `json:"gameId"`
	Modality   string          `json:"modality"`
	VotedUp    bool            `json:"votedUp"`
	Language   string          `json:"language"`
	Body       string          `json:"body"`
	ImageURI   string          `json:"imageUri"`
	GoldLabels map[string]bool `json:"goldLabels"`
}
