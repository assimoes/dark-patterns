"use client";

import { useEffect, useMemo, useState } from "react";
import { ArrowLeft, ArrowRight, Eraser, EyeOff } from "lucide-react";
import { blindReviews } from "@/lib/blind";
import { assignColors, splitBySpans } from "@/lib/highlight";
import { codebook } from "@/lib/codebook";
import { BlindReviewPane } from "./BlindReviewPane";
import { BlindPatternCard } from "./BlindPatternCard";
import { Decision } from "@/lib/types";

export function BlindAdjudicationView() {
    const reviews = blindReviews;
    const [idx, setIdx] = useState(0);
    const [decisions, setDecisions] = useState<Record<string, Decision>>({});
    const [evidence, setEvidence] = useState<Record<string, string>>({});
    const [capturing, setCapturing] = useState<string | null>(null);
    const [captureError, setCaptureError] = useState<string | null>(null);
    const [hovered, setHovered] = useState<string | null>(null);

    const review = reviews[idx];
    const keyOf = (code: string) => `${review.id}:${code}`;
    const total = codebook.length;

    // Spans to underline: only patterns the labeller marked present AND cited.
    const { segments, colors } = useMemo(() => {
        const cited = codebook
            .filter((p) => decisions[`${review.id}:${p.code}`] === "present" && evidence[`${review.id}:${p.code}`])
            .map((p) => ({ code: p.code, text: evidence[`${review.id}:${p.code}`]! }));
        const cols = assignColors(cited.map((c) => c.code));
        const spans = cited.map((c) => ({ text: c.text, code: c.code, color: cols[c.code] }));
        return { segments: splitBySpans(review.body, spans), colors: cols };
    }, [review, decisions, evidence]);

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

    // Marking absent clears that pattern's cited evidence. Side effects are
    // hoisted out of the setState updater (updaters must stay pure).
    const decide = (code: string, d: Decision) => {
        setDecisions((prev) => ({ ...prev, [keyOf(code)]: d }));
        if (d === "absent") {
            clearEvidence(code);
            if (capturing === code) cancelCapture();
        }
    };

    // Capture the current text selection as evidence, with explicit feedback when
    // it can't be used (too short, or not copied verbatim from the review).
    const captureSelection = () => {
        if (!capturing) return;
        const text = (window.getSelection()?.toString() ?? "").trim();
        if (!text) return; // a plain click, not a selection — ignore silently
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

    const go = (delta: number) => {
        cancelCapture();
        setIdx((i) => Math.min(reviews.length - 1, Math.max(0, i + delta)));
    };

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
                    Independent labelling
                </span>
                <span className="rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-xs font-medium text-sky-700">
                    panel hidden
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
            </header>

            <div className="mb-6 flex items-start gap-3 rounded-xl border border-sky-200/70 bg-sky-50/60 px-4 py-3 text-sm text-sky-900">
                <EyeOff className="mt-0.5 size-4 shrink-0 text-sky-600" />
                <p>
                    <span className="font-semibold">Independent pass.</span> The LLM panel&rsquo;s votes and evidence are
                    hidden — you label each pattern from the codebook alone. These labels form the <span className="font-medium">independent reference set</span> the
                    panel is later measured against.
                </p>
            </div>

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
                    <div className="flex items-center justify-between rounded-xl border border-slate-200 bg-white/80 px-4 py-3 backdrop-blur-sm">
                        <span className="text-sm text-slate-500">
                            Label all {total} patterns from the codebook.
                        </span>
                        <button
                            type="button"
                            onClick={resetReview}
                            className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                        >
                            <Eraser className="size-4" /> Reset
                        </button>
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
        </main>
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