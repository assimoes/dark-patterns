import type {
    AddAnnotatorInput,
    AddAnnotatorResult,
    AnnotatorRow,
    CreateGameInput,
    CreateGameResult,
    CreatePopulationInput,
    CreatePopulationResult,
    CreateRunInput,
    CreateRunResult,
    DashboardData,
    Decision,
    EnqueueAnnotationsResult,
    EnqueueScrapeInput,
    EnqueueScrapeResult,
    GameRow,
    ModelStat,
    PanelMember,
    PopulationCoverage,
    PopulationDetail,
    PopulationSummary,
    PromptRow,
    RunDetail,
    RunSummary,
    User,
} from "@/lib/types";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
    constructor(
        public status: number,
        message: string,
    ) {
        super(message);
        this.name = "ApiError";
    }
}

async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${API}${path}`, {
        credentials: 'same-origin',
        headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
        ...init,
    })

    if (!res.ok) {
        const body = await res.text().catch(() => "")
        throw new ApiError(res.status, body || res.statusText)
    }

    // no content (writes essentially)
    if (res.status === 204) {
        return undefined as T
    }

    return res.json() as Promise<T>;
}

export const api = {

    dashboard: () => fetchJSON<DashboardData>('/api/dashboard'),

    // GET /api/games/{gameId}/populations — the populations a game appears in
    gamePopulations: (gameId: string) =>
        fetchJSON<PopulationCoverage[]>(`/api/games/${gameId}/populations`),

    // GET /api/games/{gameId}/populations/{populationId}/models — distinct
    // annotated reviews per model, scoped to a single population.
    gamePopulationModels: (gameId: string, populationId: string) =>
        fetchJSON<ModelStat[]>(
            `/api/games/${gameId}/populations/${populationId}/models`,
        ),

    // GET /api/populations/{populationId}/panel — the annotators on a population.
    populationPanel: (populationId: string) =>
        fetchJSON<PanelMember[]>(`/api/populations/${populationId}/panel`),
    submitDecisions: (reviewId: string, decisions: Record<string, Decision>) =>
        fetchJSON<void>(`/api/reviews/${reviewId}/decisions`, {
            method: 'POST',
            body: JSON.stringify({ decisions })
        }),

    // POST /api/games — register a curated game. 201 Game.
    createGame: (body: CreateGameInput) =>
        fetchJSON<CreateGameResult>("/api/games", {
            method: "POST",
            body: JSON.stringify(body),
        }),
    // POST /api/annotators — add a human or llm annotator. 201.
    createAnnotator: (body: AddAnnotatorInput) =>
        fetchJSON<AddAnnotatorResult>("/api/annotators", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/populations — materialise a population from filters. 201.
    createPopulation: (body: CreatePopulationInput) =>
        fetchJSON<CreatePopulationResult>("/api/populations", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/runs — open an annotation run. 201.
    createRun: (body: CreateRunInput) =>
        fetchJSON<CreateRunResult>("/api/runs", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/scrapes — enqueue a Steam reviews scrape. 202.
    enqueueScrape: (body: EnqueueScrapeInput) =>
        fetchJSON<EnqueueScrapeResult>("/api/scrapes", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // POST /api/runs/{runId}/annotations — enqueue annotation jobs. No body. 202.
    enqueueAnnotations: (runId: string) =>
        fetchJSON<EnqueueAnnotationsResult>(`/api/runs/${runId}/annotations`, {
            method: "POST",
        }),

    // GET /api/populations — every materialised population.
    listPopulations: () => fetchJSON<PopulationSummary[]>("/api/populations"),

    // GET /api/populations/{id} — one population with per-game coverage and runs.
    getPopulation: (id: number) =>
        fetchJSON<PopulationDetail>(`/api/populations/${id}`),

    // GET /api/annotators — every annotator (human and llm).
    listAnnotators: () => fetchJSON<AnnotatorRow[]>("/api/annotators"),

    // GET /api/prompts — every prompt template.
    listPrompts: () => fetchJSON<PromptRow[]>("/api/prompts"),

    // GET /api/runs — every run.
    listRuns: () => fetchJSON<RunSummary[]>("/api/runs"),

    // GET /api/runs/{id} — one run with its panel and sample flag.
    getRun: (id: number) => fetchJSON<RunDetail>(`/api/runs/${id}`),

    // GET /api/games — every curated game with its coverage.
    listGames: () => fetchJSON<GameRow[]>("/api/games"),

}