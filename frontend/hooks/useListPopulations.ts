"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// Materialised populations
export function useListPopulations() {
    return useQuery({
        queryKey: qk.populations,
        queryFn: () => api.listPopulations(),
    });
}