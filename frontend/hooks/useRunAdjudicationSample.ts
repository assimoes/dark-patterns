"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { ApiError } from "@/lib/api/utils";
import { AdjudicationPass } from "@/lib/types";

// Persisted sample for a panel run, scoped to the pass so the decided counts match the screen.
export function useRunAdjudicationSample(runId: string, pass: AdjudicationPass, enabled: boolean) {
    return useQuery({
        queryKey: qk.adjudicationSample(runId, pass),
        queryFn: () => api.adjudication.runAdjudicationSample(runId, pass),
        enabled,
        retry: (failureCount, error) => {
            if (error instanceof ApiError && error.status === 404) return false;
            return failureCount < 3;
        },
    });
}