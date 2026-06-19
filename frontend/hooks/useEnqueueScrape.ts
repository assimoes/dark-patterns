"use client";

import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { EnqueueScrapeInput } from "@/lib/types";

// Enqueues a scrape for a game on one source.
export function useEnqueueScrape() {
    return useMutation({
        mutationFn: (body: EnqueueScrapeInput) => api.operations.enqueueScrape(body),
    });
}