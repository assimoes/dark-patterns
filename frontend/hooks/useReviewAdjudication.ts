"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// Panel read for the review currently open in the adjudication worklist
export function useReviewAdjudication(
    reviewId: string,
    panelRunId: string,
    enabled: boolean,
) {
    return useQuery({
        queryKey: qk.reviewAdjudication(reviewId, panelRunId),
        queryFn: () => api.reviewAdjudication(reviewId, panelRunId),
        enabled,
    });
}