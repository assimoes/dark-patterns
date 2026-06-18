'use client';

import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { AdjudicationPass, Decision } from "@/lib/types";
import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useSubmitDecisions(reviewId: string, runId: string, pass: AdjudicationPass) {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (decisions: Record<string, Decision>) =>
            api.adjudication.submitDecisions(reviewId, runId, pass, decisions),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.dashboard })
            // refresh this pass's worklist so the decided count moves after a save.
            queryClient.invalidateQueries({ queryKey: qk.adjudicationSample(runId, pass) })
        },
    })
}

