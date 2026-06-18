import {
    AdjudicationPass,
    AdjudicationReview,
    AdjudicationSample,
    BlindReviewData,
    CreateSampleInput,
    CreateSampleResult,
    Decision
} from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    // pass routes the save to the open or blind bucket on the same gold run, so one never overwrites the other.
    submitDecisions: (reviewId: string, runId: string, pass: AdjudicationPass, decisions: Record<string, Decision>) =>
        fetchJSON<void>(`/api/reviews/${reviewId}/decisions?run=${runId}&pass=${pass}`, {
            method: 'POST',
            body: JSON.stringify({ decisions })
        }),
    // POST /api/adjudication-samples
    createAdjudicationSample: (body: CreateSampleInput) =>
        fetchJSON<CreateSampleResult>("/api/adjudication-samples", {
            method: "POST",
            body: JSON.stringify(body),
        }),

    // GET /api/runs/{runId}/adjudication-sample?pass={pass} — pass scopes the per-review decided count.
    runAdjudicationSample: (runId: string, pass: AdjudicationPass) =>
        fetchJSON<AdjudicationSample>(`/api/runs/${runId}/adjudication-sample?pass=${pass}`),

    // GET /api/reviews/{reviewId}/adjudication?run={panelRunId}
    reviewAdjudication: (reviewId: string, panelRunId: string) =>
        fetchJSON<AdjudicationReview>(
            `/api/reviews/${reviewId}/adjudication?run=${panelRunId}`,
        ),

    // GET /api/reviews/{reviewId}/blind?run={panelRunId}
    reviewBlind: (reviewId: string, panelRunId: string) =>
        fetchJSON<BlindReviewData>(`/api/reviews/${reviewId}/blind?run=${panelRunId}`),
}