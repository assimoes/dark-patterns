"use client";

import { ThumbsDown, ThumbsUp } from "lucide-react";
import { useDashboard } from "@/hooks/useDashboard";
import { type Segment } from "@/lib/adjudication";
import type { AdjReview } from "@/lib/run";

export function ReviewPane({
    review,
    segments,
    hovered,
    onHover,
}: {
    review: AdjReview;
    segments: Segment[];
    hovered: string | null;
    onHover: (code: string | null) => void;
}) {
    // Resolve the game from the live dashboard games (keyed by external game id), not the static mock
    // list — a real review's gameId is a real external id the mock never knew, so the old static lookup
    // fell back to its first entry ("World of Tanks") for every review.
    const dashboard = useDashboard();
    const g = dashboard.data?.games.find((x) => x.id === review.gameId);
    const color = g?.color ?? "#94a3b8";
    const name = g?.name ?? review.gameId;

    return (
        <div className="rounded-2xl border border-slate-200/80 bg-white/90 shadow-sm ring-1 ring-slate-900/[0.02] lg:sticky lg:top-6">
            <div className="flex flex-wrap items-center gap-2 border-b border-slate-100 px-5 py-4">
                <span className="flex items-center gap-2 text-sm font-semibold text-slate-800">
                    <span className="size-2.5 rounded-full" style={{ backgroundColor: color }} />
                    {name}
                </span>
                <span
                    title="The review author's own Steam rating"
                    className={`ml-auto inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium ${review.votedUp
                        ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                        : "border-rose-200 bg-rose-50 text-rose-700"
                        }`}
                >
                    {review.votedUp ? <ThumbsUp className="size-3" /> : <ThumbsDown className="size-3" />}
                    {review.votedUp ? "Recommended" : "Not recommended"}
                </span>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs text-slate-500">
                    {review.language}
                </span>
                <span className="font-mono text-xs text-slate-400">#{review.id}</span>
            </div>

            <div className="px-5 py-5">
                <p className="whitespace-pre-wrap text-[15px] leading-7 text-slate-700">
                    {segments.map((s, i) =>
                        s.code ? (
                            <mark
                                key={i}
                                title={s.code}
                                aria-label={`evidence cited for ${s.code}`}
                                onMouseEnter={() => onHover(s.code!)}
                                onMouseLeave={() => onHover(null)}
                                className="cursor-help rounded-sm px-0.5 transition-colors"
                                style={{
                                    backgroundColor: hovered === s.code ? `${s.color}33` : `${s.color}14`,
                                    borderBottom: `2px solid ${s.color}`,
                                    color: "inherit",
                                }}
                            >
                                {s.text}
                            </mark>
                        ) : (
                            <span key={i}>{s.text}</span>
                        ),
                    )}
                </p>

                <p className="mt-4 border-t border-slate-100 pt-3 text-xs text-slate-400">
                    Underlines are the evidence each LLM cited, coloured by pattern. Hover a pattern card to
                    find its evidence here.
                </p>
            </div>
        </div>
    );
}