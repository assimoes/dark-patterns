"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { ApiError } from "@/lib/api/utils";

// Persisted sample for a panel run
export function useRunAdjudicationSample(runId: string, enabled: boolean) {
    return useQuery({
        queryKey: qk.adjudicationSample(runId),
        queryFn: () => api.adjudication.runAdjudicationSample(runId),
        enabled,
        retry: (failureCount, error) => {
            if (error instanceof ApiError && error.status === 404) return false;
            return failureCount < 3;
        },
    });
}