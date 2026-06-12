"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowLeft, ArrowRight, Eraser, EyeOff, Inbox, Save } from "lucide-react";
import { assignColors, splitBySpans } from "@/lib/highlight";
import { codebook } from "@/lib/codebook";
import { type Decision } from "@/lib/adjudication";
import { ApiError } from "@/lib/api";
import { useRunAdjudicationSample } from "@/hooks/useRunAdjudicationSample";
import { useReviewBlind } from "@/hooks/useReviewBlind";
import { useSubmitDecisions } from "@/hooks/useSubmitDecisions";
import type { AdjudicationSample, BlindReviewData, SampleReview } from "@/lib/types";
import { RunSelector } from "@/components/adjudicate/RunSelector";
import { BlindReviewPane } from "./BlindReviewPane";
import { BlindPatternCard } from "./BlindPatternCard";

// The blind screen, gated on a run + its sample like the adjudication screen.
// The worklist is the same persisted sample, but each review is read through the
// blind endpoint (text only — NO panel data), so the labeller scores from the
// codebook alone. Labels are persisted the same way as the panel screen.
export function BlindAdjudicationView() {
    const [runId, setRunId] = useState("");
    const sample = useRunAdjudicationSample(runId, runId !== "");

    const noSample =
        sample.isError && sample.error instanceof ApiError && sample.error.status === 404;

    return (
        <main className="mx-auto max-w-7xl px-5 py-8 sm:px-8">
            <header className="mb-5 flex flex-wrap items-center gap-4">
                <a
                    href="/"
                    className="inline-flex items-center gap-1.5 text-sm text-slate-500 transition-colors hover:text-slate-800"
                >
                    <ArrowLeft className="size-4" /> Dashboard
                </a>
                <span className="flex items-center gap-2 text-sm font-semibold text-slate-800">
                    <span className="grid size-7 place-items-center rounded-lg bg-sky-50 text-sky-600">
                        <EyeOff className="size-4" />
                    </span>
                    Blind labelling
                </span>
                <span className="rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-xs font-medium text-sky-700">
                    panel hidden
                </span>
                <div className="ml-auto">
                    <RunSelector runId={runId} onChange={setRunId} accent="sky" />
                </div>
            </header>

            <div className="mb-6 flex items-start gap-3 rounded-xl border border-sky-200/70 bg-sky-50/60 px-4 py-3 text-sm text-sky-900">
                <EyeOff className="mt-0.5 size-4 shrink-0 text-sky-600" />
                <p>
                    <span className="font-semibold">Blind pass.</span> The LLM panel&rsquo;s votes and evidence are
                    hidden — you label each pattern from the codebook alone. There is no &ldquo;accept all&rdquo;, by
                    design. These labels form the <span className="font-medium">independent reference set</span> the
                    panel is later measured against.
                </p>
            </div>

            {runId === "" ? (
                <EmptyState
                    title="Choose a panel run"
                    body="Pick a panel run above to load its sample for blind labelling."
                />
            ) : sample.isPending ? (
                <EmptyState title="Loading sample…" body="Fetching this run's worklist." />
            ) : noSample ? (
                <EmptyState
                    title="No sample for this run yet"
                    body="Draw a stratified adjudication sample before labelling."
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
                    body="This run's sample has no reviews to label."
                />
            ) : (
                <BlindWorklist sample={sample.data} runId={runId} />
            )}
        </main>
    );
}

// Mounted only with a non-empty sample so the per-review hook runs
// unconditionally. Owns the worklist position, the labeller's decisions + cited
// evidence, the selection-capture flow, and persistence.
function BlindWorklist({ sample, runId }: { sample: AdjudicationSample; runId: string }) {
    const reviews = sample.reviews;
    const [idx, setIdx] = useState(0);
    const [decisions, setDecisions] = useState<Record<string, Decision>>({});
    const [evidence, setEvidence] = useState<Record<string, string>>({});
    const [capturing, setCapturing] = useState<string | null>(null);
    const [captureError, setCaptureError] = useState<string | null>(null);
    const [hovered, setHovered] = useState<string | null>(null);

    const current: SampleReview = reviews[idx];
    const detail = useReviewBlind(current.id, true);
    const review: BlindReviewData | undefined = detail.data;

    const keyOf = (code: string) => `${current.id}:${code}`;
    const total = codebook.length;

    const submit = useSubmitDecisions(current.id, runId);

    // Spans to underline: only patterns the labeller marked present AND cited.
    const { segments, colors } = useMemo(() => {
        if (!review) return { segments: [], colors: {} as Record<string, string> };
        const cited = codebook
            .filter(
                (p) =>
                    decisions[`${current.id}:${p.code}`] === "present" &&
                    evidence[`${current.id}:${p.code}`],
            )
            .map((p) => ({ code: p.code, text: evidence[`${current.id}:${p.code}`]! }));
        const cols = assignColors(cited.map((c) => c.code));
        const spans = cited.map((c) => ({ text: c.text, code: c.code, color: cols[c.code] }));
        return { segments: splitBySpans(review.body, spans), colors: cols };
    }, [review, current.id, decisions, evidence]);

    const decided = codebook.filter((p) => decisions[keyOf(p.code)] !== undefined).length;
    const present = codebook.filter((p) => decisions[keyOf(p.code)] === "present").length;

    const clearEvidence = (code: string) =>
        setEvidence((prev) => {
            const next = { ...prev };
            delete next[keyOf(code)];
            return next;
        });

    const cancelCapture = () => {
        setCapturing(null);
        setCaptureError(null);
    };

    const decide = (code: string, d: Decision) => {
        setDecisions((prev) => ({ ...prev, [keyOf(code)]: d }));
        if (d === "absent") {
            clearEvidence(code);
            if (capturing === code) cancelCapture();
        }
    };

    const captureSelection = () => {
        if (!capturing || !review) return;
        const text = (window.getSelection()?.toString() ?? "").trim();
        if (!text) return;
        if (text.length < 3) {
            setCaptureError("Selection must be at least 3 characters.");
            return;
        }
        if (!review.body.includes(text)) {
            setCaptureError("Copy the evidence exactly from the review text.");
            return;
        }
        setEvidence((prev) => ({ ...prev, [keyOf(capturing)]: text }));
        setCapturing(null);
        setCaptureError(null);
        window.getSelection()?.removeAllRanges();
    };

    // Esc exits capture mode from anywhere on the page.
    useEffect(() => {
        if (!capturing) return;
        const onKey = (e: KeyboardEvent) => {
            if (e.key === "Escape") cancelCapture();
        };
        window.addEventListener("keydown", onKey);
        return () => window.removeEventListener("keydown", onKey);
    }, [capturing]);

    const resetReview = () => {
        setDecisions((prev) => {
            const next = { ...prev };
            for (const p of codebook) delete next[keyOf(p.code)];
            return next;
        });
        setEvidence((prev) => {
            const next = { ...prev };
            for (const p of codebook) delete next[keyOf(p.code)];
            return next;
        });
        cancelCapture();
    };

    const decisionsForReview = (): Record<string, Decision> => {
        const out: Record<string, Decision> = {};
        for (const p of codebook) {
            const d = decisions[keyOf(p.code)];
            if (d !== undefined) out[p.code] = d;
        }
        return out;
    };

    const save = () => submit.mutate(decisionsForReview());

    const go = (delta: number) => {
        cancelCapture();
        setIdx((i) => Math.min(reviews.length - 1, Math.max(0, i + delta)));
    };

    return (
        <>
            <div className="mb-6 flex flex-wrap items-center gap-3 rounded-xl border border-slate-200 bg-white/80 px-4 py-3 backdrop-blur-sm">
                <span className="rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-xs font-medium text-sky-700">
                    gold run #{sample.goldRunId}
                </span>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs text-slate-500">
                    {current.stratum}
                </span>
                <div className="ml-auto flex items-center gap-3">
                    <span className="text-sm text-slate-500 tabular-nums">
                        <span className="font-semibold text-slate-800">{decided}</span>/{total} decided
                        <span className="ml-2 text-slate-400">· {present} present</span>
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
                <EmptyState title="Loading review…" body="Fetching the review text." />
            ) : detail.isError || !review ? (
                <EmptyState
                    title="Could not load this review"
                    body="The blind read failed. Use the arrows to try another, or reselect the run."
                />
            ) : (
                <div className="grid grid-cols-1 gap-6 lg:grid-cols-5">
                    <div className="lg:col-span-2">
                        <BlindReviewPane
                            review={review}
                            segments={segments}
                            hovered={hovered}
                            capturing={capturing}
                            captureError={captureError}
                            onHover={setHovered}
                            onMouseUp={captureSelection}
                            onCancelCapture={cancelCapture}
                        />
                    </div>

                    <div className="space-y-3 lg:col-span-3">
                        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-200 bg-white/80 px-4 py-3 backdrop-blur-sm">
                            <span className="text-sm text-slate-500">
                                Label all {total} patterns from the codebook.
                            </span>
                            <div className="flex items-center gap-2">
                                <button
                                    type="button"
                                    onClick={save}
                                    disabled={submit.isPending || decided === 0}
                                    className="inline-flex items-center gap-1.5 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-sm font-medium text-emerald-700 transition-colors hover:bg-emerald-100 disabled:cursor-not-allowed disabled:opacity-50"
                                >
                                    <Save className="size-4" />
                                    {submit.isPending ? "Saving…" : "Save labels"}
                                </button>
                                <button
                                    type="button"
                                    onClick={resetReview}
                                    className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                                >
                                    <Eraser className="size-4" /> Reset
                                </button>
                            </div>
                            {submit.isError ? (
                                <span className="w-full text-xs font-medium text-rose-600">Save failed — retry.</span>
                            ) : submit.isSuccess ? (
                                <span className="w-full text-xs font-medium text-emerald-600">Saved.</span>
                            ) : null}
                        </div>

                        {codebook.map((pattern) => (
                            <BlindPatternCard
                                key={pattern.code}
                                pattern={pattern}
                                decision={decisions[keyOf(pattern.code)]}
                                evidence={evidence[keyOf(pattern.code)]}
                                capturing={capturing === pattern.code}
                                evidenceColor={colors[pattern.code]}
                                onDecide={(d) => decide(pattern.code, d)}
                                onCite={() => {
                                    setCapturing(pattern.code);
                                    setCaptureError(null);
                                }}
                                onCancelCite={() => setCapturing(null)}
                                onClearEvidence={() => clearEvidence(pattern.code)}
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
