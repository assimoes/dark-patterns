export type Comparison = {
    id: number;
    label: string;
    runAId: number;
    runBId: number;
    createdAt: string;
};

// one attribute side by side: the two runs' values and whether they match — informational context for
// reading the divergence, not a constraint.
export type AttributeDiff = {
    attribute: string;
    a: string;
    b: string;
    equal: boolean;
};

export type RunRef = { id: number; label: string; runType: string };

export type ModelAgreement = { model: string; agreedPresent: number; flips: number };

export type ComparisonDetail = Comparison & {
    runA: RunRef;
    runB: RunRef;
    attributes: AttributeDiff[];
    modelAgreement: ModelAgreement[];
    reviewsTotal: number;
    reviewsDiverged: number;
};

export type ComparisonReviewStatus = "diverged" | "converged";

export type ComparisonReviewRow = {
    reviewId: string;
    gameId: string;
    flips: number;
    status: ComparisonReviewStatus;
};

export type CreateComparisonInput = {
    label: string;
    runAId: number;
    runBId: number;
};

// GET /api/runs/{id}/distribution — one (game, pattern) detection tally: reviews where the panel
// majority flagged the code. the client rolls these up by pattern and family, per game and overall.
export type PatternCount = {
    gameId: string;
    gameName: string;
    code: string;
    reviews: number;
};
