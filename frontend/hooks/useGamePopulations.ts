"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";

// The populations a game appears in, each with its coverage. This is a lazy,
// drill-in query: the ReviewsCard only opens a game on click, so we pass
// `enabled` to hold the fetch back until the row is expanded. While disabled the
// query sits in `pending` with no request sent; flipping enabled true kicks off
// the fetch.
export function useGamePopulations(gameId: string, enabled: boolean) {
    return useQuery({
        queryKey: qk.gamePopulations(gameId),
        queryFn: () => api.gamePopulations(gameId),
        enabled,
    });
}