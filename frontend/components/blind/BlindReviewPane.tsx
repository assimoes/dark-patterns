"use client";

import { MousePointerClick, ThumbsDown, ThumbsUp } from "lucide-react";
import { useDashboard } from "@/hooks/useDashboard";
import { type BlindReview } from "@/lib/blind";
import { type Segment } from "@/lib/highlight";

export function BlindReviewPane({
    review,
    segments,
    hovered,
    capturing,
    captureError,
    onHover,
    onMouseUp,
    onCancelCapture,
}: {
    review: BlindReview;
    segments: Segment[];
    hovered: string | null;
    capturing: string | null;
    captureError: string | null;
    onHover: (code: string | null) => void;
    onMouseUp: () => void;
    onCancelCapture: () => void;
}) {
    // game from live dashboard games (by external id), not the static mock list — that fell back to "World of Tanks" for every real review.
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
                    title="The review author's own Steam rating (not a panel vote)"
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

            {capturing ? (
                <div
                    className={`flex flex-wrap items-center gap-2 border-b px-5 py-2 text-xs font-medium ${captureError ? "border-rose-100 bg-rose-50 text-rose-700" : "border-sky-100 bg-sky-50 text-sky-700"
                        }`}
                >
                    <MousePointerClick className="size-3.5" />
                    {captureError ? (
                        <span>{captureError} Select the text for <span className="font-mono">{capturing}</span> again.</span>
                    ) : (
                        <span>
                            Select the evidence text for <span className="font-mono">{capturing}</span>, then release.
                        </span>
                    )}
                    <button
                        type="button"
                        onClick={onCancelCapture}
                        className="ml-auto rounded px-1.5 py-0.5 font-semibold underline-offset-2 hover:underline"
                    >
                        Cancel (Esc)
                    </button>
                </div>
            ) : null}

            <div className="px-5 py-5">
                <p
                    onMouseUp={onMouseUp}
                    className={`whitespace-pre-wrap text-[15px] leading-7 text-slate-700 ${capturing ? "cursor-text rounded-lg bg-sky-50/40 ring-1 ring-sky-100" : ""
                        }`}
                >
                    {segments.map((s, i) =>
                        s.code ? (
                            <mark
                                key={i}
                                title={s.code}
                                aria-label={`your evidence for ${s.code}`}
                                onMouseEnter={() => onHover(s.code!)}
                                onMouseLeave={() => onHover(null)}
                                className="rounded-sm px-0.5 transition-colors"
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
                    Independent pass — no panel votes are shown. Underlines are evidence{" "}
                    <span className="font-medium text-slate-500">you</span> cited.
                </p>
            </div>
        </div>
    );
}