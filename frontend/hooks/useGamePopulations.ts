"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// The populations a game appears in, each with its coverage.
export function useGamePopulations(gameId: string, enabled: boolean) {
    return useQuery({
        queryKey: qk.gamePopulations(gameId),
        queryFn: () => api.coverage.gamePopulations(gameId),
        enabled,
    });
}