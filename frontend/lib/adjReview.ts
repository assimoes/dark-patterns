// Maps the API's panel read (AdjudicationReview) onto the AdjReview shape the
// existing adjudication logic consumes.
import type { AdjudicationReview, Decision } from "@/lib/types";
import type { AdjReview, PanelModel, PresentDetection } from "@/lib/run";
import { codebook } from "@/lib/codebook";

export type MappedAdjReview = {
    review: AdjReview;
    models: PanelModel[];
    goldDecisions: Record<string, Decision>;
};

// Derive a short display label from a model name
function shortLabel(name: string): string {
    const head = name.split(/[\s/]/)[0] ?? name;
    return head.length > 12 ? head.slice(0, 12) : head || name;
}

export function mapAdjudicationReview(data: AdjudicationReview): MappedAdjReview {
    const models: PanelModel[] = data.panelModels.map((name) => ({
        id: name,
        name,
        short: shortLabel(name),
    }));

    const known = new Set(codebook.map((p) => p.code));
    const present: PresentDetection[] = data.detections
        .filter((d) => known.has(d.patternCode))
        .map((d) => ({ code: d.patternCode, modelId: d.model, evidence: d.evidence }));

    const review: AdjReview = {
        id: data.id,
        gameId: data.gameId,
        votedUp: data.votedUp,
        language: data.language,
        body: data.body,
        present,
    };

    const goldDecisions: Record<string, Decision> = {};
    for (const [code, isPresent] of Object.entries(data.goldLabels)) {
        if (!known.has(code)) continue;
        goldDecisions[code] = isPresent ? "present" : "absent";
    }

    return { review, models, goldDecisions };
}