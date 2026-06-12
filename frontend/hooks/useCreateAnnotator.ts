"use client";

import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { AddAnnotatorInput } from "@/lib/types";

// Adds a human or llm annotator.
export function useCreateAnnotator() {
    return useMutation({
        mutationFn: (body: AddAnnotatorInput) => api.operations.createAnnotator(body),
    });
}