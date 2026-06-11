"use client";

import { useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { EnqueueScrapeInput } from "@/lib/types";

// Enqueues a Steam reviews scrape.
export function useEnqueueScrape() {
    return useMutation({
        mutationFn: (body: EnqueueScrapeInput) => api.enqueueScrape(body),
    });
}