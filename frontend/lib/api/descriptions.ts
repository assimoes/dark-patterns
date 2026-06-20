import type { DescriptionProfile, GameDescription } from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    // GET /api/games/{id}/descriptions — version history, newest first.
    list: (gameId: string) =>
        fetchJSON<GameDescription[]>(`/api/games/${gameId}/descriptions`),

    // PUT /api/game-descriptions/{id} — save a draft; the body is the edited profile. the server
    // re-renders and re-scans for valenced language.
    update: (id: number, profile: DescriptionProfile) =>
        fetchJSON<GameDescription>(`/api/game-descriptions/${id}`, {
            method: "PUT",
            body: JSON.stringify(profile),
        }),

    // POST /api/game-descriptions/{id}/approve — freeze a draft as the approved version.
    approve: (id: number) =>
        fetchJSON<GameDescription>(`/api/game-descriptions/${id}/approve`, {
            method: "POST",
        }),

    // POST /api/games/{id}/research — re-run research, producing a fresh draft.
    research: (gameId: string, disambiguationHint?: string) =>
        fetchJSON<void>(`/api/games/${gameId}/research`, {
            method: "POST",
            body: JSON.stringify({ disambiguation_hint: disambiguationHint ?? "" }),
        }),
};
