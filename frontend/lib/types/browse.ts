// GET /api/populations
export type PopulationSummary = {
    id: number;
    label: string;
    modality: string;
    individuals: number;
    createdAt: string;
};

// GET /api/populations/{id}
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

// GET /api/annotators. refCount > 0 means a run or annotation uses it, so it is frozen.
export type AnnotatorRow = {
    id: number;
    kind: string;
    label: string;
    model: string | null;
    refCount: number;
};

// GET /api/prompts. runCount > 0 means a run uses it, so it is frozen.
export type PromptRow = {
    id: number;
    name: string;
    version: number;
    modality: string;
    runCount: number;
};

// GET /api/prompts/{id} — the full prompt body for the edit form.
export type PromptDetail = {
    id: number;
    name: string;
    version: number;
    modality: string;
    system_prompt: string;
    template: string;
};

// GET /api/runs 
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

// GET /api/runs/{id}
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

// GET /api/games. artifacts > 0 means the game is frozen for delete (would orphan artifacts).
// sourceRefs maps each source to its scrape handle.
export type GameRow = {
    id: string;
    name: string;
    short: string;
    monetization: string;
    color: string;
    reviews: number;
    annotated: number;
    artifacts: number;
    sourceRefs: Record<string, string>;
};
