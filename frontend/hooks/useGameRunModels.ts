"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

export function useGameRunModels(
    gameId: string,
    runId: number,
    enabled: boolean,
) {
    return useQuery({
        queryKey: qk.gameRunModels(gameId, runId),
        queryFn: () => api.gameRunModels(gameId, runId),
        enabled,
    });
}
