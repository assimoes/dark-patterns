"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type { CreatePopulationInput } from "@/lib/types";

// Materialises a population from filter criteria.
export function useCreatePopulation() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (body: CreatePopulationInput) => api.createPopulation(body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.dashboard });
        },
    });
}