import { PopulationCoverage, ModelStat, PanelMember } from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    // GET /api/games/{gameId}/populations
    gamePopulations: (gameId: string) =>
        fetchJSON<PopulationCoverage[]>(`/api/games/${gameId}/populations`),

    // GET /api/games/{gameId}/populations/{populationId}/models
    gamePopulationModels: (gameId: string, populationId: string) =>
        fetchJSON<ModelStat[]>(
            `/api/games/${gameId}/populations/${populationId}/models`,
        ),

    gameRunModels: (gameId: string, runId: number) =>
        fetchJSON<ModelStat[]>(`/api/games/${gameId}/runs/${runId}/models`),

    // GET /api/populations/{populationId}/panel
    populationPanel: (populationId: string) =>
        fetchJSON<PanelMember[]>(`/api/populations/${populationId}/panel`),
}