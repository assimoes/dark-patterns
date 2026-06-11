"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// Get one run with its panel members and sample flag.
export function useRun(id: number | null) {
    return useQuery({
        queryKey: qk.run(id ?? 0),
        queryFn: () => api.getRun(id as number),
        enabled: id !== null,
    });
}