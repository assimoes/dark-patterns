"use client";

import { useMemo, useState } from "react";
import { ArrowLeft, ArrowRight, CheckCheck, Eraser, Layers, Save } from "lucide-react";
import { adjRun } from "@/lib/run";
import {
    highlightColors,
    majorityDecision,
    patternVotes,
    segmentReview,
    type Decision,
} from "@/lib/adjudication";

import { useSubmitDecisions } from "@/hooks/useSubmitDecisions";
import { ReviewPane } from "./ReviewPane";
import { PatternCard } from "./PatternCard";

export function AdjudicationView() {
    const reviews = adjRun.reviews;
    const [idx, setIdx] = useState(0);
    const [decisions, setDecisions] = useState<Record<string, Decision>>({});
    const [hovered, setHovered] = useState<string | null>(null);

    const review = reviews[idx];
    const votes = useMemo(() => patternVotes(review), [review]);
    const colors = useMemo(() => highlightColors(review), [review]);
    const segments = useMemo(() => segmentReview(review, colors), [review, colors]);

    const submit = useSubmitDecisions(String(review.id))

    const keyOf = (code: string) => `${review.id}:${code}`;
    const decided = votes.filter((v) => decisions[keyOf(v.pattern.code)] !== undefined).length;

    const decisionsForReview = (): Record<string, Decision> => {
        const out: Record<string, Decision> = {};
        for (const v of votes) {
            const d = decisions[keyOf(v.pattern.code)];
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

    const save = () => submit.mutate(decisionsForReview())

    return (
        <main className="mx-auto max-w-7xl px-5 py-8 sm:px-8">
            {/* top bar */}
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
                    Run #{adjRun.id}
                </span>
                <span className="rounded-full border border-violet-200 bg-violet-50 px-2 py-0.5 text-xs font-medium text-violet-700">
                    {adjRun.type}
                </span>
                <span className="text-xs text-slate-400">taxonomy v{adjRun.taxonomyVersion}</span>

                <div className="ml-auto flex items-center gap-3">
                    <span className="text-sm text-slate-500 tabular-nums">
                        <span className="font-semibold text-slate-800">{decided}</span>/{votes.length} decided
                    </span>
                    <div className="flex items-center gap-1.5">
                        <NavButton disabled={idx === 0} onClick={() => setIdx((i) => i - 1)}>
                            <ArrowLeft className="size-4" />
                        </NavButton>
                        <span className="min-w-20 text-center text-sm text-slate-600 tabular-nums">
                            Review {idx + 1} / {reviews.length}
                        </span>
                        <NavButton disabled={idx === reviews.length - 1} onClick={() => setIdx((i) => i + 1)}>
                            <ArrowRight className="size-4" />
                        </NavButton>
                    </div>
                </div>
            </header>

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-5">
                {/* left: the raw review with evidence underlined */}
                <div className="lg:col-span-2">
                    <ReviewPane review={review} segments={segments} hovered={hovered} onHover={setHovered} />
                </div>

                {/* right: toolbar + 19 pattern cards */}
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
                            decision={decisions[keyOf(pv.pattern.code)]}
                            evidenceColor={colors[pv.pattern.code]}
                            onDecide={(d) => decide(pv.pattern.code, d)}
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

