import type {
    DashboardData,
    Decision,
    ModelStat,
    PanelMember,
    PopulationCoverage,
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
    // GET /api/games/{gameId}/populations — the populations a game appears in,
    // each with its reviews/annotated coverage.
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
}