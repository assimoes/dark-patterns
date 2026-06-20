import {
    AnnotatorRow,
    GameRow,
    PopulationDetail,
    PopulationSummary,
    PromptDetail,
    PromptRow,
    RunDetail,
    RunSummary
} from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    // GET /api/populations
    listPopulations: () => fetchJSON<PopulationSummary[]>("/api/populations"),

    // GET /api/populations/{id}
    getPopulation: (id: number) =>
        fetchJSON<PopulationDetail>(`/api/populations/${id}`),

    // GET /api/annotators
    listAnnotators: () => fetchJSON<AnnotatorRow[]>("/api/annotators"),

    // GET /api/prompts
    listPrompts: () => fetchJSON<PromptRow[]>("/api/prompts"),

    // GET /api/prompts/{id}
    getPrompt: (id: number) => fetchJSON<PromptDetail>(`/api/prompts/${id}`),

    // GET /api/runs
    listRuns: () => fetchJSON<RunSummary[]>("/api/runs"),

    // GET /api/runs/{id}
    getRun: (id: number) => fetchJSON<RunDetail>(`/api/runs/${id}`),

    // GET /api/games
    listGames: () => fetchJSON<GameRow[]>("/api/games"),
}