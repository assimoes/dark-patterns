"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// All prompt templates.
export function useListPrompts() {
    return useQuery({
        queryKey: qk.prompts,
        queryFn: () => api.listPrompts(),
    });
}