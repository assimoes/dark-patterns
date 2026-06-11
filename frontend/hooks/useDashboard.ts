'use client';

import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { useQuery } from "@tanstack/react-query";

export function useDashboard() {
    return useQuery({
        queryKey: qk.dashboard,
        queryFn: () => api.dashboard(),
    })
}

