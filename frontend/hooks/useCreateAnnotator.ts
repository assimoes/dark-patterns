"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type { AddAnnotatorInput } from "@/lib/types";

// Adds a human or llm annotator.
export function useCreateAnnotator() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (body: AddAnnotatorInput) => api.operations.createAnnotator(body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.annotators });
        },
    });
}