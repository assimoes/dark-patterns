import {
    AddAnnotatorInput,
    AddAnnotatorResult,
    CreateGameInput,
    CreateGameResult,
    CreatePopulationInput,
    CreatePopulationResult,
    CreateRunInput,
    CreateRunResult,
    EnqueueAnnotationsResult,
    EnqueueScrapeInput,
    EnqueueScrapeResult
} from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    // POST /api/games 
    createGame: (body: CreateGameInput) =>
        fetchJSON<CreateGameResult>("/api/games", {
            method: "POST",
            body: JSON.stringify(body),
        }),
    // POST /api/annotators
    createAnnotator: (body: AddAnnotatorInput) =>
        fetchJSON<AddAnnotatorResult>("/api/annotators", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/populations
    createPopulation: (body: CreatePopulationInput) =>
        fetchJSON<CreatePopulationResult>("/api/populations", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/runs
    createRun: (body: CreateRunInput) =>
        fetchJSON<CreateRunResult>("/api/runs", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/scrapes
    enqueueScrape: (body: EnqueueScrapeInput) =>
        fetchJSON<EnqueueScrapeResult>("/api/scrapes", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/runs/{runId}/annotations
    enqueueAnnotations: (runId: string) =>
        fetchJSON<EnqueueAnnotationsResult>(`/api/runs/${runId}/annotations`, {
            method: "POST",
        }),
}