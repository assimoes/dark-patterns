export type Monetization = 'f2p' | 'premium'

// Same shape as the GO API
export type Game = {
    id: string;
    name: string;
    short: string;
    monetization: string;
    color: string;
}

export type PopulationStat = {
    gameId: string;
    individuals: number;
}

export type Kind = 'llm' | 'human'

export type Member = {
    kind: Kind;
    label: string;
}

export type Run = {
    id: string;
    label: string;
    population: string;
    createdAt: string;
    members: Member[];
}

export type ReviewStat = {
    gameId: string;
    runId: number;
    prompt: string;
    reviews: number;
    annotated: number;
}

// Payload for GET /api/dashboard.
export type DashboardData = {
    games: Game[];
    populations: PopulationStat[];
    runs: Run[];
    reviewStats: ReviewStat[];
};


// The logged-in user, returned by /api/login and /api/me.
export type User = { id: number; email: string; displayName: string };

// A single adjudication call the auditor records for a pattern. This is the
// value written to POST /api/reviews/{reviewId}/decisions.
export type Decision = "present" | "absent";


// Per-LLM-model completed annotation count, from
// GET /api/games/{gameId}/populations/{populationId}/models — distinct annotated
// reviews per model WITHIN a single population. `model` is the display name.
export type ModelStat = { model: string; annotated: number };

// One population a game appears in, with its coverage, from
// GET /api/games/{gameId}/populations. `populationId` is numeric; `label` is the
// population's display name; reviews/annotated scope the coverage bar.
export type PopulationCoverage = {
    populationId: number;
    label: string;
    reviews: number;
    annotated: number;
};

// One member of a population's annotation panel, from
// GET /api/populations/{populationId}/panel. `kind` is "llm" or "human";
// `label` is the annotator's display name.
export type PanelMember = { kind: string; label: string };

// POST /api/games — register a curated Steam game.
export type CreateGameInput = {
    external_game_id: number;
    name?: string;
    short?: string;
    monetization?: Monetization;
    color?: string;
};
export type CreateGameResult = {
    id: number;
    name: string;
    short: string;
    monetization: Monetization;
    color: string;
};


// POST /api/annotators — add a human or an llm panel member.
export type AnnotatorKind = "human" | "llm";
export type AddAnnotatorInput = {
    kind: AnnotatorKind;
    label: string;
    family?: string;
    slug?: string;
    name?: string;
    modalities?: string[];
};
export type AddAnnotatorResult = {
    id: number;
    kind: AnnotatorKind;
    label: string;
    model_id: number | null;
};

// POST /api/populations — materialise a population from filter criteria.
export type CreatePopulationInput = {
    description?: string;
    min_hours_played?: number;
    per_game_cap?: number;
    artifacts_cutoff?: string; // RFC3339
};
export type CreatePopulationResult = {
    population_id: number;
    inserted_individuals: number;
    total_individuals: number;
};


// POST /api/runs — open an annotation run over a population with a prompt.
export type RunType = "llm_panel" | "gold";
export type CreateRunInput = {
    population_id: number;
    prompt_id: number;
    taxonomy_version?: number;
    run_type?: RunType;
    temperature?: number;
    annotator_ids?: number[];
};
export type CreateRunResult = {
    run_id: number;
    run_type: RunType;
    population_id: number;
    prompt_id: number;
    taxonomy_version: number;
    annotator_ids: number[];
};

// POST /api/scrapes — enqueue a Steam reviews scrape for an app id.
export type ScrapeFilter = "recent" | "updated";
export type EnqueueScrapeInput = {
    app: string;
    filter?: ScrapeFilter;
    lang?: string;
    max?: number;
};
export type EnqueueScrapeResult = {
    enqueued: true;
    game_id: number;
    filter: string;
    language: string;
    max: number;
};

// POST /api/runs/{runId}/annotations — enqueue annotation jobs for a run.
export type EnqueueAnnotationsResult = {
    enqueued: number;
    run_id: number;
};


// Browse-area read types

// GET /api/populations
export type PopulationSummary = {
    id: number;
    label: string;
    modality: string;
    individuals: number;
    createdAt: string;
};

// GET /api/populations/{id} — a population with its per-game coverage and the
// runs opened over it.
export type PopulationDetail = {
    id: number;
    label: string;
    modality: string;
    createdAt: string;
    perGame: {
        gameId: string;
        name: string;
        reviews: number;
        annotated: number;
    }[];
    runs: {
        id: number;
        label: string;
        runType: string;
    }[];
};

// GET /api/annotators — one row per annotator.
export type AnnotatorRow = {
    id: number;
    kind: string;
    label: string;
    model: string | null;
};

// GET /api/prompts — one row per prompt template, with its version and the
// modality it targets.
export type PromptRow = {
    id: number;
    name: string;
    version: number;
    modality: string;
};

// GET /api/runs — one row per run.

export type RunSummary = {
    id: number;
    runType: string;
    label: string;
    population: string;
    populationId: number;
    promptId: number;
    taxonomyVersion: number | null;
    createdAt: string;
    panelSize: number;
};

// GET /api/runs/{id} — a run with its panel members and whether an adjudication
export type RunDetail = {
    id: number;
    runType: string;
    population: string;
    populationId: number;
    promptId: number;
    taxonomyVersion: number | null;
    createdAt: string;
    panel: {
        kind: string;
        label: string;
    }[];
    hasSample: boolean;
};

// GET /api/games — one row per curated game
export type GameRow = {
    id: string;
    name: string;
    short: string;
    monetization: string;
    color: string;
    reviews: number;
    annotated: number;
};

// Mock data

export const games: Game[] = [
    { id: "wot", name: "World of Tanks", short: "WoT", monetization: "f2p", color: "#f59e0b" },
    { id: "poe", name: "Path of Exile", short: "PoE", monetization: "f2p", color: "#f43f5e" },
    { id: "cod", name: "Call of Duty: IW", short: "CoD-IW", monetization: "premium", color: "#0ea5e9" },
    { id: "archeage", name: "ArcheAge", short: "ArcheAge", monetization: "f2p", color: "#10b981" },
];

export const gameById = (id: string): Game =>
    games.find((g) => g.id === id) ?? games[0];

// Format a number with thousands separators, e.g. 12345 -> "12,345".
export const fmt = (n: number) => n.toLocaleString("en-US");



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

// Positive panel detection, from GET /api/reviews/{reviewId}/adjudication
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

// Blind read for one review, from GET /api/reviews/{reviewId}/blind
export type BlindReviewData = {
    id: string;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
};