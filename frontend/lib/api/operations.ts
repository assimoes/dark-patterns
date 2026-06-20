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
    EnqueueScrapeResult,
    ImpactResult,
    PromptInput,
    UpdateAnnotatorInput,
    UpdateGameInput,
    UploadImagesInput,
    UploadImagesResult
} from "@/lib/types";
import type { PromptRow } from "@/lib/types/browse";
import { fetchJSON } from "@/lib/api/utils";

// reads a File into raw base64 (no data: prefix). the backend wraps it into a data uri itself.
async function fileToBase64(file: File): Promise<string> {
    const bytes = new Uint8Array(await file.arrayBuffer());
    let binary = "";
    for (let i = 0; i < bytes.length; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
}

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

    uploadImages: async (body: UploadImagesInput) => {
        const images = await Promise.all(
            body.images.map(async (img) => ({
                data: await fileToBase64(img.file),
                mimeType: img.file.type || "image/png",
                description: img.description ?? "",
            })),
        );

        return fetchJSON<UploadImagesResult>("/api/images", {
            method: "POST",
            body: JSON.stringify({ game_id: body.game_id, images }),
        });
    },

    // PUT /api/games/{id}
    updateGame: (id: string, body: UpdateGameInput) =>
        fetchJSON<CreateGameResult>(`/api/games/${id}`, {
            method: "PUT",
            body: JSON.stringify(body),
        }),
    // DELETE /api/games/{id}
    deleteGame: (id: string) =>
        fetchJSON<void>(`/api/games/${id}`, { method: "DELETE" }),

    // POST /api/prompts
    createPrompt: (body: PromptInput) =>
        fetchJSON<PromptRow>("/api/prompts", {
            method: "POST",
            body: JSON.stringify(body),
        }),
    // PUT /api/prompts/{id}
    updatePrompt: (id: string, body: PromptInput) =>
        fetchJSON<PromptRow>(`/api/prompts/${id}`, {
            method: "PUT",
            body: JSON.stringify(body),
        }),
    // DELETE /api/prompts/{id}
    deletePrompt: (id: string) =>
        fetchJSON<void>(`/api/prompts/${id}`, { method: "DELETE" }),

    // PUT /api/annotators/{id}
    updateAnnotator: (id: string, body: UpdateAnnotatorInput) =>
        fetchJSON<void>(`/api/annotators/${id}`, {
            method: "PUT",
            body: JSON.stringify(body),
        }),
    // DELETE /api/annotators/{id}
    deleteAnnotator: (id: string) =>
        fetchJSON<void>(`/api/annotators/${id}`, { method: "DELETE" }),

    // DELETE /api/populations/{id}
    deletePopulation: (id: string) =>
        fetchJSON<void>(`/api/populations/${id}`, { method: "DELETE" }),
    // GET /api/populations/{id}/impact
    populationImpact: (id: string) =>
        fetchJSON<ImpactResult>(`/api/populations/${id}/impact`),

    // DELETE /api/runs/{id}
    deleteRun: (id: string) =>
        fetchJSON<void>(`/api/runs/${id}`, { method: "DELETE" }),
    // GET /api/runs/{id}/impact
    runImpact: (id: string) =>
        fetchJSON<ImpactResult>(`/api/runs/${id}/impact`),
}