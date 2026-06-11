'use client';

import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { Decision } from "@/lib/types";
import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useSubmitDecisions(reviewId: string) {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (decisions: Record<string, Decision>) =>
            api.submitDecisions(reviewId, decisions),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.dashboard })
        },
    })
}

