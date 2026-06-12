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

// GET /api/annotators
export type AnnotatorRow = {
    id: number;
    kind: string;
    label: string;
    model: string | null;
};

// GET /api/prompts
export type PromptRow = {
    id: number;
    name: string;
    version: number;
    modality: string;
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

// GET /api/games
export type GameRow = {
    id: string;
    name: string;
    short: string;
    monetization: string;
    color: string;
    reviews: number;
    annotated: number;
};
