"use client";

import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { CreateSampleInput } from "@/lib/types";

// Draws a stratified adjudication sample from a panel run
export function useCreateAdjudicationSample() {
    return useMutation({
        mutationFn: (body: CreateSampleInput) => api.adjudication.createAdjudicationSample(body),
    });
}