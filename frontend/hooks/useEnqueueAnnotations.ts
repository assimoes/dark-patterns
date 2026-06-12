"use client";

import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";

// Enqueues annotation jobs for an existing run.
export function useEnqueueAnnotations() {
    return useMutation({
        mutationFn: (runId: string) => api.operations.enqueueAnnotations(runId),
    });
}