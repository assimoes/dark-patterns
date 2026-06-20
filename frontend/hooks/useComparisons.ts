"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type { CreateComparisonInput } from "@/lib/types";

export function useComparisons() {
    return useQuery({ queryKey: qk.comparisons, queryFn: () => api.comparisons.list() });
}

export function useComparison(id: number) {
    return useQuery({ queryKey: qk.comparison(id), queryFn: () => api.comparisons.get(id) });
}

export function useComparisonReviews(id: number) {
    return useQuery({ queryKey: qk.comparisonReviews(id), queryFn: () => api.comparisons.reviews(id) });
}

export function useCreateComparison() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (body: CreateComparisonInput) => api.comparisons.create(body),
        onSuccess: () => qc.invalidateQueries({ queryKey: qk.comparisons }),
    });
}

export function useDeleteComparison() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => api.comparisons.remove(id),
        onSuccess: () => qc.invalidateQueries({ queryKey: qk.comparisons }),
    });
}
