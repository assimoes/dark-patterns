"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// All annotators (human and llm). 
export function useListAnnotators() {
    return useQuery({
        queryKey: qk.annotators,
        queryFn: () => api.browse.listAnnotators(),
    });
}