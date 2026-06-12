"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// One population with its per-game coverage and runs.
export function usePopulation(id: number | null) {
    return useQuery({
        queryKey: qk.population(id ?? 0),
        queryFn: () => api.browse.getPopulation(id as number),
        enabled: id !== null,
    });
}