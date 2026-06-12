"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
    ArrowLeft,
    ArrowRight,
    CheckCheck,
    Eraser,
    Inbox,
    Layers,
    Save,
} from "lucide-react";
import {
    highlightColors,
    majorityDecision,
    patternVotes,
    segmentReview,
    type Decision,
} from "@/lib/adjudication";
import { mapAdjudicationReview } from "@/lib/adjReview";
import { ApiError } from "@/lib/api/utils";
import { useRunAdjudicationSample } from "@/hooks/useRunAdjudicationSample";
import { useReviewAdjudication } from "@/hooks/useReviewAdjudication";
import { useSubmitDecisions } from "@/hooks/useSubmitDecisions";
import type { AdjudicationSample, SampleReview } from "@/lib/types";
import { RunSelector } from "./RunSelector";
import { ReviewPane } from "./ReviewPane";
import { PatternCard } from "./PatternCard";

// The adjudication screen, now gated on a run + its persisted sample. Choosing a
// panel run loads that run's sample (the worklist); the per-review panel votes
// are then fetched on demand as the auditor pages through it. The pattern cards
// and review pane are unchanged — only the data source moved to the API.
export function AdjudicationView() {
    const [runId, setRunId] = useState("");
    const sample = useRunAdjudicationSample(runId, runId !== "");

    const noSample =
        sample.isError && sample.error instanceof ApiError && sample.error.status === 404;

    return (
        <main className="mx-auto max-w-7xl px-5 py-8 sm:px-8">
            <header className="mb-6 flex flex-wrap items-center gap-4">
                <a
                    href="/"
                    className="inline-flex items-center gap-1.5 text-sm text-slate-500 transition-colors hover:text-slate-800"
                >
                    <ArrowLeft className="size-4" /> Dashboard
                </a>
                <span className="flex items-center gap-2 text-sm font-semibold text-slate-800">
                    <span className="grid size-7 place-items-center rounded-lg bg-violet-50 text-violet-600">
                        <Layers className="size-4" />
                    </span>
                    Adjudication
                </span>
                <div className="ml-auto">
                    <RunSelector runId={runId} onChange={setRunId} accent="violet" />
                </div>
            </header>

            {runId === "" ? (
                <EmptyState
                    title="Choose a panel run"
                    body="Pick a panel run above to load its adjudication sample."
                />
            ) : sample.isPending ? (
                <EmptyState title="Loading sample…" body="Fetching this run's worklist." />
            ) : noSample ? (
                <EmptyState
                    title="No sample for this run yet"
                    body="Draw a stratified adjudication sample before reviewing."
                    action={
                        <Link
                            href="/operations/adjudication-sample"
                            className="mt-3 inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700"
                        >
                            Create adjudication sample
                        </Link>
                    }
                />
            ) : sample.isError ? (
                <EmptyState
                    title="Could not load the sample"
                    body="The request failed. Try selecting the run again."
                />
            ) : sample.data.reviews.length === 0 ? (
                <EmptyState
                    title="The sample is empty"
                    body="This run's sample has no reviews to adjudicate."
                />
            ) : (
                <SampleWorklist runId={runId} sample={sample.data} />
            )}
        </main>
    );
}

// Mounted only once a non-empty sample exists, so the per-review hooks below run
// unconditionally. Holds the worklist position, the per-review-and-code decision
// map, and persists each review's calls via useSubmitDecisions.
function SampleWorklist({
    runId,
    sample,
}: {
    runId: string;
    sample: AdjudicationSample;
}) {
    const reviews = sample.reviews;
    const [idx, setIdx] = useState(0);
    const [decisions, setDecisions] = useState<Record<string, Decision>>({});
    const [hovered, setHovered] = useState<string | null>(null);

    const current: SampleReview = reviews[idx];
    const detail = useReviewAdjudication(current.id, runId, true);

    // Map the wire response into the AdjReview the existing logic consumes, plus
    // the real panel models and any gold labels to pre-fill from.
    const mapped = useMemo(
        () => (detail.data ? mapAdjudicationReview(detail.data) : null),
        [detail.data],
    );

    const votes = useMemo(
        () => (mapped ? patternVotes(mapped.review, mapped.models) : []),
        [mapped],
    );
    const colors = useMemo(
        () => (mapped ? highlightColors(mapped.review) : {}),
        [mapped],
    );
    const segments = useMemo(
        () => (mapped ? segmentReview(mapped.review, colors) : []),
        [mapped, colors],
    );

    // The reviewId the write endpoint expects is the individual id string; runId scopes the write to
    // the same panel run (and so the same gold run + taxonomy version) the screen is reading.
    const submit = useSubmitDecisions(current.id, runId);

    const keyOf = (code: string) => `${current.id}:${code}`;

    // Effective decision for a pattern: the auditor's local call if they made one,
    // otherwise the gold label the response pre-filled (so a re-opened review shows
    // its saved state without forcing every cell to be touched again).
    const effective = (code: string): Decision | undefined =>
        decisions[keyOf(code)] ?? mapped?.goldDecisions[code];

    const decided = votes.filter((v) => effective(v.pattern.code) !== undefined).length;

    const decisionsForReview = (): Record<string, Decision> => {
        const out: Record<string, Decision> = {};
        for (const v of votes) {
            const d = effective(v.pattern.code);
            if (d !== undefined) out[v.pattern.code] = d;
        }
        return out;
    };

    const decide = (code: string, d: Decision) =>
        setDecisions((prev) => ({ ...prev, [keyOf(code)]: d }));

    const acceptAll = () =>
        setDecisions((prev) => {
            const next = { ...prev };
            for (const v of votes) next[keyOf(v.pattern.code)] = majorityDecision(v.verdict);
            return next;
        });

    const clearAll = () =>
        setDecisions((prev) => {
            const next = { ...prev };
            for (const v of votes) delete next[keyOf(v.pattern.code)];
            return next;
        });

    const save = () => submit.mutate(decisionsForReview());

    const go = (delta: number) =>
        setIdx((i) => Math.min(reviews.length - 1, Math.max(0, i + delta)));

    return (
        <>
            <div className="mb-6 flex flex-wrap items-center gap-3 rounded-xl border border-slate-200 bg-white/80 px-4 py-3 backdrop-blur-sm">
                <span className="rounded-full border border-violet-200 bg-violet-50 px-2 py-0.5 text-xs font-medium text-violet-700">
                    gold run #{sample.goldRunId}
                </span>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs text-slate-500">
                    {current.stratum}
                </span>
                <div className="ml-auto flex items-center gap-3">
                    <span className="text-sm text-slate-500 tabular-nums">
                        <span className="font-semibold text-slate-800">{decided}</span>/{votes.length} decided
                    </span>
                    <div className="flex items-center gap-1.5">
                        <NavButton disabled={idx === 0} onClick={() => go(-1)}>
                            <ArrowLeft className="size-4" />
                        </NavButton>
                        <span className="min-w-20 text-center text-sm text-slate-600 tabular-nums">
                            Review {idx + 1} / {reviews.length}
                        </span>
                        <NavButton disabled={idx === reviews.length - 1} onClick={() => go(1)}>
                            <ArrowRight className="size-4" />
                        </NavButton>
                    </div>
                </div>
            </div>

            {detail.isPending ? (
                <EmptyState title="Loading review…" body="Fetching the panel votes for this review." />
            ) : detail.isError ? (
                <EmptyState
                    title="Could not load this review"
                    body="The panel read failed. Use the arrows to try another, or reselect the run."
                />
            ) : !mapped ? null : (
                <div className="grid grid-cols-1 gap-6 lg:grid-cols-5">
                    <div className="lg:col-span-2">
                        <ReviewPane
                            review={mapped.review}
                            segments={segments}
                            hovered={hovered}
                            onHover={setHovered}
                        />
                    </div>

                    <div className="space-y-3 lg:col-span-3">
                        <div className="flex flex-wrap items-center gap-3 rounded-xl border border-slate-200 bg-white/80 px-4 py-3 backdrop-blur-sm">
                            <button
                                type="button"
                                onClick={acceptAll}
                                className="inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-slate-700"
                            >
                                <CheckCheck className="size-4" /> Accept all (majority)
                            </button>
                            <button
                                type="button"
                                onClick={clearAll}
                                className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                            >
                                <Eraser className="size-4" /> Clear
                            </button>
                            <button
                                type="button"
                                onClick={save}
                                disabled={submit.isPending || decided === 0}
                                className="inline-flex items-center gap-1.5 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-sm font-medium text-emerald-700 transition-colors hover:bg-emerald-100 disabled:cursor-not-allowed disabled:opacity-50"
                            >
                                <Save className="size-4" />
                                {submit.isPending ? "Saving…" : "Save decisions"}
                            </button>
                            {submit.isError ? (
                                <span className="text-xs font-medium text-rose-600">Save failed — retry.</span>
                            ) : submit.isSuccess ? (
                                <span className="text-xs font-medium text-emerald-600">Saved.</span>
                            ) : null}
                            <Legend />
                        </div>

                        {votes.map((pv) => (
                            <PatternCard
                                key={pv.pattern.code}
                                pv={pv}
                                decision={effective(pv.pattern.code)}
                                evidenceColor={colors[pv.pattern.code]}
                                onDecide={(d) => decide(pv.pattern.code, d)}
                                onHover={setHovered}
                            />
                        ))}
                    </div>
                </div>
            )}
        </>
    );
}

function EmptyState({
    title,
    body,
    action,
}: {
    title: string;
    body: string;
    action?: React.ReactNode;
}) {
    return (
        <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-white/60 px-6 py-16 text-center">
            <span className="grid size-11 place-items-center rounded-xl bg-slate-100 text-slate-400">
                <Inbox className="size-5" />
            </span>
            <h2 className="mt-3 text-sm font-semibold text-slate-800">{title}</h2>
            <p className="mt-1 max-w-sm text-sm text-slate-500">{body}</p>
            {action}
        </div>
    );
}

function NavButton({
    disabled,
    onClick,
    children,
}: {
    disabled: boolean;
    onClick: () => void;
    children: React.ReactNode;
}) {
    return (
        <button
            type="button"
            disabled={disabled}
            onClick={onClick}
            className="grid size-8 place-items-center rounded-lg border border-slate-200 bg-white text-slate-600 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
        >
            {children}
        </button>
    );
}

function Legend() {
    const items = [
        { c: "bg-emerald-400", t: "majority" },
        { c: "bg-amber-400", t: "minority / tie" },
        { c: "bg-slate-300", t: "none" },
        { c: "bg-sky-400", t: "override" },
    ];
    return (
        <div className="ml-auto flex flex-wrap items-center gap-x-3 gap-y-1">
            {items.map((it) => (
                <span key={it.t} className="flex items-center gap-1.5 text-xs text-slate-500">
                    <span className={`size-2.5 rounded-full ${it.c}`} />
                    {it.t}
                </span>
            ))}
        </div>
    );
}
