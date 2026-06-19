import { Monetization } from "./dashboard";

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
    artifacts_cutoff?: string; // RFC3339 timestamp
    game_ids?: number[];
    modality?: "text" | "image"
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
