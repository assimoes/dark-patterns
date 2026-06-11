'use client';

import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { useQuery } from "@tanstack/react-query";

export function useGameModels(gameId: string, enabled: boolean) {
    return useQuery({
        queryKey: qk.gameModels(gameId),
        queryFn: () => api.gameModels(gameId),
        enabled,
    })
}
