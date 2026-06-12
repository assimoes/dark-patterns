import {
    AdjudicationReview,
    AdjudicationSample,
    BlindReviewData,
    CreateSampleInput,
    CreateSampleResult,
    Decision
} from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    submitDecisions: (reviewId: string, runId: string, decisions: Record<string, Decision>) =>
        fetchJSON<void>(`/api/reviews/${reviewId}/decisions?run=${runId}`, {
            method: 'POST',
            body: JSON.stringify({ decisions })
        }),
    // POST /api/adjudication-samples
    createAdjudicationSample: (body: CreateSampleInput) =>
        fetchJSON<CreateSampleResult>("/api/adjudication-samples", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // GET /api/runs/{runId}/adjudication-sample
    runAdjudicationSample: (runId: string) =>
        fetchJSON<AdjudicationSample>(`/api/runs/${runId}/adjudication-sample`),

    // GET /api/reviews/{reviewId}/adjudication?run={panelRunId}
    reviewAdjudication: (reviewId: string, panelRunId: string) =>
        fetchJSON<AdjudicationReview>(
            `/api/reviews/${reviewId}/adjudication?run=${panelRunId}`,
        ),

    // GET /api/reviews/{reviewId}/blind
    reviewBlind: (reviewId: string) =>
        fetchJSON<BlindReviewData>(`/api/reviews/${reviewId}/blind`),
}