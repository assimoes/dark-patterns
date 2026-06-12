"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// All runs
export function useListRuns() {
    return useQuery({
        queryKey: qk.runs,
        queryFn: () => api.browse.listRuns(),
    });
}