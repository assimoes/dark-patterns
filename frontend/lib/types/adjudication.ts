// A single adjudication call the auditor records for a pattern
export type Decision = "present" | "absent";

// Which adjudication pass a review is labelled in: open sees the panel, blind hides it. Both live on the
// same gold run as separate rows, so neither overwrites the other.
export type AdjudicationPass = "open" | "blind";

// POST /api/adjudication-samples 
export type CreateSampleInput = {
    panelRunId: number;
    flaggedMajority: number;
    flaggedSplit: number;
    silent: number;
    seed?: number;
};

export type CreateSampleResult = {
    sampleId: number;
    goldRunId: number;
    seed: number;
    counts: {
        flagged_majority: number;
        flagged_split: number;
        silent: number;
    };
};

// GET /api/runs/{runId}/adjudication-sample
export type SampleReview = {
    id: string;
    gameId: string;
    stratum: string;
    votedUp: boolean;
    language: string;
    decided: number;
};

// The whole sample for a run
export type AdjudicationSample = {
    sampleId: number;
    goldRunId: number;
    reviews: SampleReview[];
};

// Positive panel detection
export type Detection = {
    patternCode: string;
    model: string;
    evidence: string;
    explanation: string;
};

// Full panel read for one review
export type AdjudicationReview = {
    id: string;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
    panelModels: string[];
    detections: Detection[];
    goldLabels: Record<string, boolean>;
};

// Blind read for one review, from GET /api/reviews/{reviewId}/blind. panel votes stay hidden, but the
// auditors own saved labels come back so a reopened review keeps its decisions.
export type BlindReviewData = {
    id: string;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
    goldLabels: Record<string, boolean>;
};