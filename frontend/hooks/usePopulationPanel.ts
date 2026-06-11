"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// The annotation panel (llm models and human annotators) for a population.
// Lazy like its sibling: gated on `enabled` so it only fetches once the
// population row is expanded.
export function usePopulationPanel(populationId: string, enabled: boolean) {
    return useQuery({
        queryKey: qk.populationPanel(populationId),
        queryFn: () => api.populationPanel(populationId),
        enabled,
    });
}