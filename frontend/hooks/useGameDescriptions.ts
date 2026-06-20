"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type { DescriptionProfile } from "@/lib/types";

// useGameDescriptions loads a game's description version history.
export function useGameDescriptions(gameId: string, enabled = true) {
    return useQuery({
        queryKey: qk.gameDescriptions(gameId),
        queryFn: () => api.descriptions.list(gameId),
        enabled: enabled && gameId !== "",
    });
}

// after a write, refresh the game's descriptions and the games list (the badge).
function useRefresh(gameId: string) {
    const qc = useQueryClient();
    return () => {
        qc.invalidateQueries({ queryKey: qk.gameDescriptions(gameId) });
        qc.invalidateQueries({ queryKey: qk.games });
    };
}

export function useUpdateDescription(gameId: string) {
    const done = useRefresh(gameId);
    return useMutation({
        mutationFn: ({ id, profile }: { id: number; profile: DescriptionProfile }) =>
            api.descriptions.update(id, profile),
        onSuccess: done,
    });
}

export function useApproveDescription(gameId: string) {
    const done = useRefresh(gameId);
    return useMutation({
        mutationFn: (id: number) => api.descriptions.approve(id),
        onSuccess: done,
    });
}

export function useResearchGame(gameId: string) {
    const done = useRefresh(gameId);
    return useMutation({
        mutationFn: (hint?: string) => api.descriptions.research(gameId, hint),
        onSuccess: done,
    });
}
