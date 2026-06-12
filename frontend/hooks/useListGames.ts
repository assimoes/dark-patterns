"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// All curated games with its coverage.
export function useListGames() {
    return useQuery({
        queryKey: qk.games,
        queryFn: () => api.browse.listGames(),
    });
}