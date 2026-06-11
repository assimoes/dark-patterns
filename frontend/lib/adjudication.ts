import { codebook, type Pattern } from "@/lib/codebook";
import { panelModels, type AdjReview } from "@/lib/run";
import { assignColors, splitBySpans, type Segment } from "@/lib/highlight";

import type { Decision } from "@/lib/types";

export type { Segment };
export type { Decision };

export type Verdict = "present" | "absent" | "tie";
export type Signal = "green" | "yellow" | "gray";

export type Vote = { modelId: string; short: string; present: boolean; evidence?: string };

export type PatternVote = {
    pattern: Pattern;
    votes: Vote[];
    presentCount: number;
    verdict: Verdict;
    signal: Signal;
    evidence: string[]; // unique cited spans, present votes only
};

// Tally the 4 panel votes for every pattern in a review.
export function patternVotes(review: AdjReview): PatternVote[] {
    return codebook.map((pattern) => {
        const votes: Vote[] = panelModels.map((m) => {
            const d = review.present.find((x) => x.code === pattern.code && x.modelId === m.id);
            return { modelId: m.id, short: m.short, present: Boolean(d), evidence: d?.evidence };
        });
        const presentCount = votes.filter((v) => v.present).length;
        const verdict: Verdict =
            presentCount >= 3 ? "present" : presentCount <= 1 ? "absent" : "tie";
        const signal: Signal =
            presentCount >= 3 ? "green" : presentCount === 0 ? "gray" : "yellow";
        const evidence = Array.from(
            new Set(votes.filter((v) => v.present && v.evidence).map((v) => v.evidence as string)),
        );
        return { pattern, votes, presentCount, verdict, signal, evidence };
    });
}

// An override is a decision that contradicts a decisive panel majority.
export function isOverride(verdict: Verdict, decision?: Decision): boolean {
    return verdict !== "tie" && decision !== undefined && decision !== verdict;
}

// What "Accept all (majority)" sets a pattern to. Ties default to absent.
export function majorityDecision(verdict: Verdict): Decision {
    return verdict === "present" ? "present" : "absent";
}

// ---- review highlighting (shared algorithm in lib/highlight.ts) ----

// One colour per pattern that has any detection in this review.
export function highlightColors(review: AdjReview): Record<string, string> {
    return assignColors(review.present.map((d) => d.code));
}

// Split the body into plain + highlighted segments at the cited evidence spans.
export function segmentReview(review: AdjReview, colors: Record<string, string>): Segment[] {
    const seen = new Set<string>();
    const spans = [] as { text: string; code: string; color: string }[];
    for (const d of review.present) {
        if (seen.has(d.evidence)) continue;
        seen.add(d.evidence);
        spans.push({ text: d.evidence, code: d.code, color: colors[d.code] });
    }
    return splitBySpans(review.body, spans);
}