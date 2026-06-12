"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// Per-model annotation counts scoped to a single population — distinct annotated
// reviews per model. A second-level lazy drill-in: gated on `enabled` so nothing
// is fetched until the auditor expands the population row inside an open game.
export function useGamePopulationModels(
    gameId: string,
    populationId: string,
    enabled: boolean,
) {
    return useQuery({
        queryKey: qk.gamePopulationModels(gameId, populationId),
        queryFn: () => api.coverage.gamePopulationModels(gameId, populationId),
        enabled,
    });
}