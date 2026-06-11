// Centralised query-key factory.
export const qk = {
    dashboard: ["dashboard"] as const,
    gamePopulations: (gameId: string) => ["gamePopulations", gameId] as const,
    gamePopulationModels: (gameId: string, populationId: string) =>
        ["gamePopulationModels", gameId, populationId] as const,
    populationPanel: (populationId: string) => ["populationPanel", populationId] as const,
};