"use client";

import { Fragment, useMemo, useState } from "react";
import { useReviewAdjudication } from "@/hooks/useReviewAdjudication";
import { assignColors, splitBySpans } from "@/lib/highlight";
import { codebook } from "@/lib/codebook";
import type { AdjudicationReview, Detection } from "@/lib/types";

const nameByCode = new Map(codebook.map((p) => [p.code, p.name]));

function shortLabel(model: string) {
    const tail = model.split("/").pop() ?? model;
    return tail.length > 18 ? tail.slice(0, 18) : tail;
}

// a column in the matrix: one model under one run. the key pins it for the focus feature.
type Col = { run: "A" | "B"; model: string; key: string };

// ComparisonReview fetches the same review under both runs and lays the panels out as a matrix: patterns
// down the side, each run's models across the top. a cell is green (present) or red (absent). click a
// pattern to see each model's evidence/explanation; focus models to pin a few columns side by side.
export function ComparisonReview({
    reviewId,
    runAId,
    runBId,
    runALabel,
    runBLabel,
}: {
    reviewId: string;
    runAId: number;
    runBId: number;
    runALabel: string;
    runBLabel: string;
}) {
    const a = useReviewAdjudication(reviewId, String(runAId), true);
    const b = useReviewAdjudication(reviewId, String(runBId), true);

    if (a.isPending || b.isPending) return <p className="text-sm text-slate-400">Loading review…</p>;
    if (a.isError || b.isError || !a.data || !b.data) return <p className="text-sm text-rose-600">Could not load the review for both runs.</p>;

    return <Matrix a={a.data} b={b.data} runALabel={runALabel} runBLabel={runBLabel} />;
}

function Matrix({ a, b, runALabel, runBLabel }: { a: AdjudicationReview; b: AdjudicationReview; runALabel: string; runBLabel: string }) {
    const [expanded, setExpanded] = useState<string | null>(null);
    const [focus, setFocus] = useState<Set<string>>(new Set());

    const cols: Col[] = useMemo(
        () => [
            ...a.panelModels.map((m) => ({ run: "A" as const, model: m, key: `A:${m}` })),
            ...b.panelModels.map((m) => ({ run: "B" as const, model: m, key: `B:${m}` })),
        ],
        [a.panelModels, b.panelModels],
    );

    const visible = focus.size === 0 ? cols : cols.filter((c) => focus.has(c.key));
    const visA = visible.filter((c) => c.run === "A");
    const visB = visible.filter((c) => c.run === "B");

    const sharedModels = useMemo(() => {
        const sb = new Set(b.panelModels);
        return new Set(a.panelModels.filter((m) => sb.has(m)));
    }, [a.panelModels, b.panelModels]);

    const detFor = (review: AdjudicationReview, code: string, model: string) =>
        review.detections.find((d) => d.patternCode === code && d.model === model);

    const present = (col: Col, code: string) =>
        Boolean(detFor(col.run === "A" ? a : b, code, col.model));

    const codes = useMemo(() => {
        const set = new Set<string>();
        for (const d of [...a.detections, ...b.detections]) set.add(d.patternCode);
        return [...set];
    }, [a.detections, b.detections]);

    // colors identify models, not patterns: a model's column header and its cited evidence in the body
    // share one colour, so an underline points back to the model that wrote it.
    const modelColors = useMemo(() => {
        const slugs = [...new Set([...a.panelModels, ...b.panelModels])].sort();
        return assignColors(slugs);
    }, [a.panelModels, b.panelModels]);

    const segments = useMemo(() => {
        const spans = [...a.detections, ...b.detections]
            .filter((d) => d.evidence)
            .map((d) => ({ text: d.evidence, code: d.model, color: modelColors[d.model] ?? "#94a3b8" }));
        return splitBySpans(a.body, spans);
    }, [a.detections, b.detections, a.body, modelColors]);

    const patterns = useMemo(() => {
        return codes
            .map((code) => {
                const flips = [...sharedModels].filter(
                    (m) => Boolean(detFor(a, code, m)) !== Boolean(detFor(b, code, m)),
                ).length;
                return { code, flips };
            })
            .sort((x, y) => y.flips - x.flips || x.code.localeCompare(y.code));
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [codes, sharedModels, a.detections, b.detections]);

    const toggleFocus = (key: string) =>
        setFocus((prev) => {
            const next = new Set(prev);
            next.has(key) ? next.delete(key) : next.add(key);
            return next;
        });

    const colCount = 1 + visA.length + visB.length;

    return (
        <div className="flex flex-col gap-4">
            <div className="rounded-xl border border-slate-200 bg-white p-4 text-sm leading-relaxed text-slate-700">
                {segments.map((s, i) =>
                    s.code ? (
                        <mark key={i} style={{ backgroundColor: `${s.color}1f`, borderBottom: `2px solid ${s.color}` }} className="rounded px-0.5">
                            {s.text}
                        </mark>
                    ) : (
                        <Fragment key={i}>{s.text}</Fragment>
                    ),
                )}
            </div>

            <div className="flex flex-col gap-2">
                <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-400">
                    <span><Dot present /> present</span>
                    <span><Dot present={false} /> absent</span>
                    <span>· select models to focus a few side by side · click a pattern row to expand</span>
                    {focus.size > 0 ? (
                        <button type="button" onClick={() => setFocus(new Set())} className="rounded-lg border border-slate-200 px-2 py-0.5 font-medium text-slate-600 hover:bg-slate-50">
                            show all
                        </button>
                    ) : null}
                </div>
                <div className="flex flex-wrap gap-x-6 gap-y-2">
                    <FocusGroup label={runALabel} cols={cols.filter((c) => c.run === "A")} focus={focus} colors={modelColors} onToggle={toggleFocus} />
                    <FocusGroup label={runBLabel} cols={cols.filter((c) => c.run === "B")} focus={focus} colors={modelColors} onToggle={toggleFocus} />
                </div>
            </div>

            <div className="overflow-x-auto rounded-xl border border-slate-200">
                <table className="w-full border-collapse text-sm">
                    <thead>
                        <tr className="border-b border-slate-100">
                            <th rowSpan={2} className="px-3 py-2 text-left align-bottom text-xs font-medium text-slate-400">Pattern</th>
                            {visA.length > 0 ? (
                                <th colSpan={visA.length} className="border-l border-slate-100 px-2 py-1.5 text-center text-xs font-semibold text-slate-600">
                                    {runALabel}
                                </th>
                            ) : null}
                            {visB.length > 0 ? (
                                <th colSpan={visB.length} className="border-l border-slate-100 px-2 py-1.5 text-center text-xs font-semibold text-slate-600">
                                    {runBLabel}
                                </th>
                            ) : null}
                        </tr>
                        <tr className="border-b border-slate-100">
                            {visA.map((c) => (
                                <ModelHeader key={c.key} col={c} color={modelColors[c.model]} leftBorder />
                            ))}
                            {visB.map((c, i) => (
                                <ModelHeader key={c.key} col={c} color={modelColors[c.model]} leftBorder={i === 0} />
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {patterns.length === 0 ? (
                            <tr>
                                <td colSpan={colCount} className="px-3 py-4 text-center text-sm text-slate-400">No model flagged any pattern in either run.</td>
                            </tr>
                        ) : (
                            patterns.map((p) => (
                                <Fragment key={p.code}>
                                    <tr
                                        onClick={() => setExpanded((e) => (e === p.code ? null : p.code))}
                                        className={`cursor-pointer border-b border-slate-50 ${p.flips > 0 ? "bg-rose-50/40" : ""} hover:bg-slate-50`}
                                    >
                                        <td className="px-3 py-2">
                                            <span className="flex items-center gap-2 text-xs">
                                                <span className="font-medium text-slate-700">{p.code}</span>
                                                <span className="text-slate-400">{nameByCode.get(p.code) ?? ""}</span>
                                                {p.flips > 0 ? <span className="rounded-full bg-rose-100 px-1.5 text-[10px] font-medium text-rose-700">{p.flips}</span> : null}
                                            </span>
                                        </td>
                                        {visA.map((c) => (
                                            <td key={c.key} className="border-l border-slate-50 px-2 py-2 text-center"><Dot present={present(c, p.code)} /></td>
                                        ))}
                                        {visB.map((c, i) => (
                                            <td key={c.key} className={`px-2 py-2 text-center ${i === 0 ? "border-l border-slate-100" : "border-l border-slate-50"}`}><Dot present={present(c, p.code)} /></td>
                                        ))}
                                    </tr>
                                    {expanded === p.code ? (
                                        <tr className="border-b border-slate-100 bg-slate-50/40">
                                            <td colSpan={colCount} className="px-3 py-3">
                                                <Evidence code={p.code} cols={visible} a={a} b={b} detFor={detFor} colors={modelColors} />
                                            </td>
                                        </tr>
                                    ) : null}
                                </Fragment>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

function ModelHeader({ col, color, leftBorder }: { col: Col; color?: string; leftBorder: boolean }) {
    return (
        <th className={`px-2 py-1.5 ${leftBorder ? "border-l border-slate-100" : ""}`}>
            <div className="flex flex-col items-center gap-1" title={col.model}>
                <span className="max-w-[8rem] truncate text-[11px] font-medium text-slate-600">{shortLabel(col.model)}</span>
                <span className="h-1 w-8 rounded-full" style={{ backgroundColor: color ?? "#94a3b8" }} />
            </div>
        </th>
    );
}

function FocusGroup({ label, cols, focus, colors, onToggle }: { label: string; cols: Col[]; focus: Set<string>; colors: Record<string, string>; onToggle: (key: string) => void }) {
    if (cols.length === 0) return null;
    return (
        <div className="flex flex-col gap-1">
            <span className="text-[11px] font-semibold text-slate-500">{label}</span>
            <div className="flex flex-wrap gap-1.5">
                {cols.map((c) => {
                    const on = focus.has(c.key);
                    return (
                        <button
                            key={c.key}
                            type="button"
                            onClick={() => onToggle(c.key)}
                            className={`inline-flex items-center gap-1.5 rounded-lg border px-2 py-1 text-[11px] font-medium transition-colors ${on ? "border-slate-900 bg-slate-900 text-white" : "border-slate-200 text-slate-600 hover:bg-slate-50"}`}
                        >
                            <span className="size-2 rounded-full" style={{ backgroundColor: colors[c.model] ?? "#94a3b8" }} />
                            {shortLabel(c.model)}
                        </button>
                    );
                })}
            </div>
        </div>
    );
}

function Dot({ present }: { present: boolean }) {
    return present ? (
        <span title="present" className="inline-block size-2.5 rounded-full bg-emerald-500 align-middle" />
    ) : (
        <span title="absent" className="inline-block text-rose-400 align-middle leading-none">✕</span>
    );
}

function Evidence({ code, cols, a, b, detFor, colors }: { code: string; cols: Col[]; a: AdjudicationReview; b: AdjudicationReview; detFor: (r: AdjudicationReview, code: string, model: string) => Detection | undefined; colors: Record<string, string> }) {
    const present = cols
        .map((c) => ({ col: c, det: detFor(c.run === "A" ? a : b, code, c.model) }))
        .filter((x) => x.det);

    if (present.length === 0) return <p className="text-xs text-slate-400">No visible model flagged {code}.</p>;

    return (
        <div className="grid grid-cols-1 gap-2 md:grid-cols-2">
            {present.map(({ col, det }) => (
                <div key={col.key} className="rounded-lg border border-slate-100 bg-white p-2.5 text-xs">
                    <p className="mb-1 flex items-center gap-1.5 font-semibold text-slate-600">
                        <span className="size-2 rounded-full" style={{ backgroundColor: colors[col.model] ?? "#94a3b8" }} />
                        <span className="rounded bg-slate-100 px-1 py-0.5 text-[10px] text-slate-500">Run {col.run}</span>
                        {shortLabel(col.model)}
                    </p>
                    {det!.evidence ? <p className="mb-1 italic text-slate-600">“{det!.evidence}”</p> : null}
                    <p className="text-slate-600">{det!.explanation || "—"}</p>
                </div>
            ))}
        </div>
    );
}
