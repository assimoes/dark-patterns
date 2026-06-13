"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// distinct annotated reviews per model for one population.
// gated on `enabled` — lazy drill-in, fetches only once the population row is expanded
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