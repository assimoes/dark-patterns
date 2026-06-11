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