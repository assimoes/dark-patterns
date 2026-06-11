export const qk = {
    dashboard: ["dashboard"] as const,
    gameModels: (gameId: string) => ['gameModels', gameId] as const,
}

