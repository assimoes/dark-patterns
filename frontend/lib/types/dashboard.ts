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

export type Monetization = 'f2p' | 'premium'
