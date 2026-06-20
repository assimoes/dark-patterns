import { Monetization } from "./dashboard";

// POST /api/games — register a curated game. the id is auto-assigned; source_refs maps each source to
// its scrape handle, e.g. {"steam":"570","reddit":"r/EVE"}.
export type CreateGameInput = {
    name: string;
    short?: string;
    monetization?: Monetization;
    color?: string;
    source_refs?: Record<string, string>;
};
export type CreateGameResult = {
    id: number;
    name: string;
    short: string;
    monetization: Monetization;
    color: string;
    sourceRefs: Record<string, string>;
};

// PUT /api/games/{id} — edit a game's display fields and source handles.
export type UpdateGameInput = {
    name: string;
    short: string;
    monetization: Monetization;
    color: string;
    source_refs: Record<string, string>;
};

// POST /api/prompts and PUT /api/prompts/{id}.
export type PromptInput = {
    name: string;
    version: number;
    modality: string;
    system_prompt?: string;
    template: string;
};

// PUT /api/annotators/{id} — only the label is editable.
export type UpdateAnnotatorInput = {
    label: string;
};

// GET /api/{populations,runs}/{id}/impact — the blast radius of a cascade delete.
export type ImpactResult = {
    individuals: number;
    runs: number;
    annotations: number;
    samples: number;
    adjudications: number;
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

// POST /api/scrapes — enqueue a scrape for a game on one source. source is whatever key the game has in
// its source_refs; the backend resolves the target from there (a target here overrides it). filter and
// lang are optional knobs a source may use.
export type ScrapeFilter = "recent" | "updated";

export type EnqueueScrapeInput = {
    game_id: number;
    source: string;
    target?: string;
    filter?: ScrapeFilter;
    lang?: string;
    max?: number;
};
export type EnqueueScrapeResult = {
    enqueued: true;
    game_id: number;
    source: string;
    target: string;
    max: number;
};

// POST /api/runs/{runId}/annotations — enqueue annotation jobs for a run.
export type EnqueueAnnotationsResult = {
    enqueued: number;
    run_id: number;
};
