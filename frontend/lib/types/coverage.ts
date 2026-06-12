// GET /api/games/{gameId}/populations/{populationId}/models
export type ModelStat = { model: string; annotated: number };

// GET /api/games/{gameId}/populations
export type PopulationCoverage = {
    populationId: number;
    label: string;
    reviews: number;
    annotated: number;
};

// GET /api/populations/{populationId}/panel. `kind` is "llm" or "human"
export type PanelMember = { kind: string; label: string };
