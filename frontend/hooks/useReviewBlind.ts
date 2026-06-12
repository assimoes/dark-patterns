"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// Blind read for the review currently open in the blind worklist 
export function useReviewBlind(reviewId: string, enabled: boolean) {
    return useQuery({
        queryKey: qk.reviewBlind(reviewId),
        queryFn: () => api.reviewBlind(reviewId),
        enabled,
    });
}