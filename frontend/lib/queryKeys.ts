// Centralised query-key factory.
export const qk = {
    dashboard: ["dashboard"] as const,
    gamePopulations: (gameId: string) => ["gamePopulations", gameId] as const,
    gamePopulationModels: (gameId: string, populationId: string) =>
        ["gamePopulationModels", gameId, populationId] as const,
    gameRunModels: (gameId: string, runId: number) =>
        ["gameRunModels", gameId, runId] as const,
    populationPanel: (populationId: string) => ["populationPanel", populationId] as const,
    populations: ["populations"] as const,
    population: (id: number) => ["population", id] as const,
    annotators: ["annotators"] as const,
    prompts: ["prompts"] as const,
    runs: ["runs"] as const,
    run: (id: number) => ["run", id] as const,
    games: ["games"] as const,
};