import type {
    Comparison,
    ComparisonDetail,
    ComparisonReviewRow,
    CreateComparisonInput,
} from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    // GET /api/comparisons
    list: () => fetchJSON<Comparison[]>("/api/comparisons"),

    // GET /api/comparisons/{id}
    get: (id: number) => fetchJSON<ComparisonDetail>(`/api/comparisons/${id}`),

    // GET /api/comparisons/{id}/reviews — the divergence worklist.
    reviews: (id: number) => fetchJSON<ComparisonReviewRow[]>(`/api/comparisons/${id}/reviews`),

    // POST /api/comparisons
    create: (body: CreateComparisonInput) =>
        fetchJSON<Comparison>("/api/comparisons", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // DELETE /api/comparisons/{id}
    remove: (id: number) => fetchJSON<void>(`/api/comparisons/${id}`, { method: "DELETE" }),
};
