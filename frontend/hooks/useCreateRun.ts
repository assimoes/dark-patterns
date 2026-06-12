"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type { CreateRunInput } from "@/lib/types";

// Opens an annotation run over a population.
export function useCreateRun() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (body: CreateRunInput) => api.operations.createRun(body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.dashboard });
        },
    });
}