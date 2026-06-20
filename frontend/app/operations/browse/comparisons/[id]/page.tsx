"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowRight } from "lucide-react";
import { BrowsePanel, TableError, TableSkeleton } from "@/components/operations/BrowseTable";
import { ComparisonReview } from "@/components/operations/ComparisonReview";
import { useComparison, useComparisonReviews } from "@/hooks/useComparisons";
import type { AttributeDiff, ComparisonReviewRow, ModelAgreement } from "@/lib/types";

export default function ComparisonDetailPage() {
    const params = useParams<{ id: string }>();
    const id = Number(params.id);

    const detail = useComparison(id);
    const reviews = useComparisonReviews(id);

    const [selected, setSelected] = useState<ComparisonReviewRow | null>(null);
    const [divergentOnly, setDivergentOnly] = useState(true);

    return (
        // break out of the operations max-w-3xl shell so the matrix has room.
        <div className="relative left-1/2 w-screen -translate-x-1/2 px-5 sm:px-8">
        <div className="mx-auto flex max-w-7xl flex-col gap-6">
            <Link href="/operations/browse/comparisons" className="text-sm font-medium text-slate-500 hover:text-slate-700">
                ← All comparisons
            </Link>

            <BrowsePanel
                title={detail.data ? detail.data.label : "Comparison"}
                subtitle={
                    detail.data
                        ? `${detail.data.runA.label}  ↔  ${detail.data.runB.label}`
                        : "Loading…"
                }
            >
                {detail.isPending ? (
                    <TableSkeleton columns={3} />
                ) : detail.isError || !detail.data ? (
                    <TableError onRetry={() => detail.refetch()} />
                ) : (
                    <div className="flex flex-col gap-5">
                        <RunDiffTable attributes={detail.data.attributes} />

                        <div className="grid grid-cols-3 gap-3">
                            <Stat label="Shared reviews" value={detail.data.reviewsTotal} />
                            <Stat label="Diverged" value={detail.data.reviewsDiverged} tone="rose" />
                            <Stat label="Converged" value={detail.data.reviewsTotal - detail.data.reviewsDiverged} tone="emerald" />
                        </div>

                        <ModelAgreementTable agreement={detail.data.modelAgreement} />
                    </div>
                )}
            </BrowsePanel>

            <BrowsePanel
                title="Reviews"
                subtitle="Sorted by flips. Click one to compare below."
                badge={
                    <button
                        type="button"
                        onClick={() => setDivergentOnly((v) => !v)}
                        className={`rounded-lg border px-2.5 py-1 text-xs font-medium transition-colors ${divergentOnly ? "border-slate-900 bg-slate-900 text-white" : "border-slate-200 text-slate-600 hover:bg-slate-50"}`}
                    >
                        diverged only
                    </button>
                }
            >
                {reviews.isPending ? (
                    <TableSkeleton columns={2} />
                ) : reviews.isError ? (
                    <TableError onRetry={() => reviews.refetch()} />
                ) : (
                    <ReviewChips
                        rows={(reviews.data ?? []).filter((r) => !divergentOnly || r.status === "diverged")}
                        selectedId={selected?.reviewId ?? null}
                        onSelect={setSelected}
                    />
                )}
            </BrowsePanel>

            <BrowsePanel
                title={selected ? `Review ${selected.reviewId}` : "Review"}
                subtitle={selected ? `${selected.flips} flip${selected.flips === 1 ? "" : "s"} · game ${selected.gameId}` : "Pick a review above to compare."}
            >
                {selected && detail.data ? (
                    <ComparisonReview
                        reviewId={selected.reviewId}
                        runAId={detail.data.runA.id}
                        runBId={detail.data.runB.id}
                        runALabel={detail.data.runA.label}
                        runBLabel={detail.data.runB.label}
                    />
                ) : (
                    <p className="text-sm text-slate-400">Select a review above.</p>
                )}
            </BrowsePanel>
        </div>
        </div>
    );
}

function RunDiffTable({ attributes }: { attributes: AttributeDiff[] }) {
    return (
        <div>
            <h3 className="mb-2 text-xs font-medium text-slate-400">How the runs differ</h3>
            <div className="overflow-hidden rounded-xl border border-slate-100">
                {attributes.map((at) => (
                    <div key={at.attribute} className="grid grid-cols-[8rem_1fr_auto_1fr] items-center gap-2 border-b border-slate-50 px-3 py-2 text-xs last:border-0">
                        <span className="font-medium capitalize text-slate-600">{at.attribute}</span>
                        <span className="truncate font-mono text-slate-500" title={at.a}>{at.a}</span>
                        <span className={at.equal ? "text-emerald-600" : "text-amber-600"}>{at.equal ? "=" : "≠"}</span>
                        <span className="truncate font-mono text-slate-500" title={at.b}>{at.b}</span>
                    </div>
                ))}
            </div>
        </div>
    );
}

function ModelAgreementTable({ agreement }: { agreement: ModelAgreement[] }) {
    const [open, setOpen] = useState<string | null>(null);
    if (agreement.length === 0) return null;
    return (
        <div>
            <h3 className="mb-2 text-xs font-medium text-slate-400">
                Per-model agreement
                <span className="ml-1 font-normal text-slate-300">— of the patterns each model flagged, how many stayed the same vs changed between the two runs. Click a row for the counts.</span>
            </h3>
            <div className="flex flex-col gap-0.5">
                {agreement.map((m) => {
                    const union = m.agreedPresent + m.flips;
                    const same = union === 0 ? 0 : Math.round((m.agreedPresent / union) * 100);
                    const changed = union === 0 ? 0 : 100 - same;
                    const isOpen = open === m.model;
                    return (
                        <div key={m.model} className="rounded-lg hover:bg-slate-50">
                            <button
                                type="button"
                                onClick={() => setOpen((o) => (o === m.model ? null : m.model))}
                                className="flex w-full items-center gap-3 px-2 py-1.5 text-left text-xs"
                            >
                                <span className="w-48 shrink-0 truncate font-medium text-slate-600" title={m.model}>{m.model}</span>
                                <div className="flex h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                                    <div className="h-full bg-emerald-400" style={{ width: `${same}%` }} />
                                    <div className="h-full bg-rose-300" style={{ width: `${changed}%` }} />
                                </div>
                                <span className="w-44 shrink-0 text-right tabular-nums">
                                    <span className="font-medium text-emerald-600">{same}% same</span>
                                    <span className="text-slate-300"> · </span>
                                    <span className="font-medium text-rose-500">{changed}% changed</span>
                                </span>
                            </button>
                            {isOpen ? (
                                <div className="grid grid-cols-3 gap-2 px-2 pb-2">
                                    <Mini label="Flagged in both runs" value={m.agreedPresent} tone="emerald" />
                                    <Mini label="Changed (one run only)" value={m.flips} tone="rose" />
                                    <Mini label="Total patterns flagged" value={union} />
                                </div>
                            ) : null}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

function Mini({ label, value, tone }: { label: string; value: number; tone?: "emerald" | "rose" }) {
    const valueCls = tone === "emerald" ? "text-emerald-600" : tone === "rose" ? "text-rose-600" : "text-slate-700";
    return (
        <div className="rounded-lg border border-slate-100 bg-white px-3 py-2">
            <p className="text-[11px] text-slate-400">{label}</p>
            <p className={`text-lg font-semibold tabular-nums ${valueCls}`}>{value}</p>
        </div>
    );
}

function ReviewChips({ rows, selectedId, onSelect }: { rows: ComparisonReviewRow[]; selectedId: string | null; onSelect: (r: ComparisonReviewRow) => void }) {
    if (rows.length === 0) return <p className="px-1 py-3 text-sm text-slate-400">No reviews.</p>;
    return (
        <div className="flex max-h-64 flex-wrap gap-1.5 overflow-y-auto">
            {rows.map((r) => (
                <button
                    key={r.reviewId}
                    type="button"
                    onClick={() => onSelect(r)}
                    title={`game ${r.gameId} · ${r.flips} flip${r.flips === 1 ? "" : "s"}`}
                    className={`inline-flex items-center gap-1.5 rounded-lg border px-2 py-1 text-xs transition-colors ${selectedId === r.reviewId ? "border-slate-900 bg-slate-900 text-white" : "border-slate-200 text-slate-600 hover:bg-slate-50"}`}
                >
                    <StatusDot status={r.status} />
                    <span className="font-mono">#{r.reviewId}</span>
                    <span className={`tabular-nums ${selectedId === r.reviewId ? "text-slate-300" : "text-slate-400"}`}>{r.flips}</span>
                </button>
            ))}
        </div>
    );
}

function StatusDot({ status }: { status: ComparisonReviewRow["status"] }) {
    const cls = status === "diverged" ? "bg-rose-400" : "bg-emerald-400";
    return <span className={`size-2 rounded-full ${cls}`} title={status} />;
}

function Stat({ label, value, tone }: { label: string; value: number; tone?: "rose" | "amber" | "emerald" }) {
    const valueCls = tone === "rose" ? "text-rose-600" : tone === "amber" ? "text-amber-600" : tone === "emerald" ? "text-emerald-600" : "text-slate-800";
    return (
        <div className="rounded-xl border border-slate-100 bg-slate-50/50 px-4 py-3">
            <p className="text-xs font-medium text-slate-400">{label}</p>
            <p className={`mt-0.5 text-2xl font-semibold tabular-nums ${valueCls}`}>{value}</p>
        </div>
    );
}
