"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type {
    PromptInput,
    UpdateAnnotatorInput,
    UpdateGameInput,
} from "@/lib/types";

// invalidate the keys a list page reads after a write, so the table refreshes.
function useInvalidate(keys: readonly (readonly unknown[])[]) {
    const qc = useQueryClient();
    return () => keys.forEach((key) => qc.invalidateQueries({ queryKey: key }));
}

export function useUpdateGame() {
    const done = useInvalidate([qk.games, qk.dashboard]);
    return useMutation({
        mutationFn: ({ id, body }: { id: string; body: UpdateGameInput }) =>
            api.operations.updateGame(id, body),
        onSuccess: done,
    });
}

export function useDeleteGame() {
    const done = useInvalidate([qk.games, qk.dashboard]);
    return useMutation({
        mutationFn: (id: string) => api.operations.deleteGame(id),
        onSuccess: done,
    });
}

export function useCreatePrompt() {
    const done = useInvalidate([qk.prompts]);
    return useMutation({
        mutationFn: (body: PromptInput) => api.operations.createPrompt(body),
        onSuccess: done,
    });
}

export function useUpdatePrompt() {
    const done = useInvalidate([qk.prompts]);
    return useMutation({
        mutationFn: ({ id, body }: { id: string; body: PromptInput }) =>
            api.operations.updatePrompt(id, body),
        onSuccess: done,
    });
}

export function useDeletePrompt() {
    const done = useInvalidate([qk.prompts]);
    return useMutation({
        mutationFn: (id: string) => api.operations.deletePrompt(id),
        onSuccess: done,
    });
}

export function useUpdateAnnotator() {
    const done = useInvalidate([qk.annotators]);
    return useMutation({
        mutationFn: ({ id, body }: { id: string; body: UpdateAnnotatorInput }) =>
            api.operations.updateAnnotator(id, body),
        onSuccess: done,
    });
}

export function useDeleteAnnotator() {
    const done = useInvalidate([qk.annotators]);
    return useMutation({
        mutationFn: (id: string) => api.operations.deleteAnnotator(id),
        onSuccess: done,
    });
}

export function useDeletePopulation() {
    const done = useInvalidate([qk.populations, qk.dashboard]);
    return useMutation({
        mutationFn: (id: string) => api.operations.deletePopulation(id),
        onSuccess: done,
    });
}

export function useDeleteRun() {
    const done = useInvalidate([qk.runs, qk.dashboard]);
    return useMutation({
        mutationFn: (id: string) => api.operations.deleteRun(id),
        onSuccess: done,
    });
}
