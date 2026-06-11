"use client";

import { useState } from "react";
import { ChevronRight, MessageSquare } from "lucide-react";
import { Card, Badge } from "@/components/ui/Card";
import { fmt, gameById } from "@/lib/types";
import type { ReviewStat, Game } from '@/lib/types'
import { useGameModels } from '@/hooks/useGameModels'

type ReviewsCardArgs = {
    games: Game[];
    reviewStats: ReviewStat[];
}

export function ReviewsCard({ games, reviewStats }: ReviewsCardArgs) {

    // Track which games are expanded. A Set lets several stay open at once.
    const [open, setOpen] = useState<Set<string>>(new Set());

    const toggle = (id: string) =>
        setOpen((prev) => {
            const next = new Set(prev);
            next.has(id) ? next.delete(id) : next.add(id);
            return next;
        });

    const totalReviews = reviewStats.reduce((s, r) => s + r.reviews, 0)
    const totalAnnotated = reviewStats.reduce((s, r) => s + r.annotated, 0)
    const totalPct = totalReviews === 0 ? 0 : Math.round((totalAnnotated / totalReviews) * 100)
    return (
        <Card
            icon={<MessageSquare className="size-5" />}
            iconClass="bg-emerald-50 text-emerald-600"
            title="Reviews & annotation"
            subtitle="Coverage per game — drill into models"
            badge={<Badge className="border-emerald-200 bg-emerald-50 text-emerald-700">{totalPct}% annotated</Badge>}
        >
            <div className="mb-5 grid grid-cols-3 gap-3">
                <Summary label="Reviews" value={fmt(totalReviews)} />
                <Summary label="Annotated" value={fmt(totalAnnotated)} accent />
                <Summary label="Coverage" value={`${totalPct}%`} />
            </div>

            <ul className="divide-y divide-slate-100 overflow-hidden rounded-xl border border-slate-100">
                {reviewStats.map((s) => (
                    <GameRow
                        key={s.gameId}
                        stat={s}
                        game={gameById(s.gameId)}
                        isOpen={open.has(s.gameId)}
                        onToggle={() => toggle(s.gameId)}
                    />
                ))}
            </ul>
        </Card>
    );
}

type GameRowArgs = {
    stat: ReviewStat;
    game: Game | undefined;
    isOpen: boolean;
    onToggle: () => void;
}

function GameRow({ stat, game, isOpen, onToggle }: GameRowArgs) {
    const models = useGameModels(stat.gameId, isOpen)

    const color = game?.color ?? '#94a3b8';
    const name = game?.name ?? stat.gameId;
    const pct = stat.reviews === 0 ? 0 : Math.round((stat.annotated / stat.reviews) * 100)


    return (
        <li>
            <button
                type="button"
                onClick={onToggle}
                aria-expanded={isOpen}
                className="flex w-full items-center gap-3 px-4 py-3 text-left text-sm transition-colors hover:bg-slate-50"
            >
                <ChevronRight
                    className={`size-4 shrink-0 text-slate-400 transition-transform duration-200 ${isOpen ? "rotate-90" : ""
                        }`}
                />
                <span className="flex items-center gap-2 font-medium text-slate-800">
                    <span className="size-2.5 shrink-0 rounded-full" style={{ backgroundColor: color }} />
                    {name}
                </span>

                <span className="ml-auto hidden text-slate-500 tabular-nums sm:inline">
                    <span className="font-medium text-slate-900">{fmt(stat.annotated)}</span> /{" "}
                    {fmt(stat.reviews)}
                </span>

                <div className="hidden h-1.5 w-24 overflow-hidden rounded-full bg-slate-100 md:block">
                    <div
                        className="h-full rounded-full"
                        style={{ width: `${pct}%`, backgroundColor: color }}
                    />
                </div>

                <span className="w-10 text-right text-xs font-semibold text-slate-700 tabular-nums">
                    {pct}%
                </span>
            </button>

            {isOpen ? (
                <div className="space-y-2.5 border-t border-slate-100 bg-slate-50/60 px-4 py-3.5">
                    <p className="text-[11px] font-medium uppercase tracking-wide text-slate-400">
                        Annotated by model
                    </p>

                    {models.isPending ? (
                        <p className="text-xs text-slate-400">Loading model coverage…</p>
                    ) : models.isError ? (
                        <p className="text-xs text-rose-500">Couldn&apos;t load model coverage.</p>
                    ) : (
                        models.data.map((pm) => {
                            const modelPct = stat.reviews === 0 ? 0 : Math.round((pm.annotated / stat.reviews) * 100);
                            return (
                                <div key={pm.model} className="flex items-center gap-3 text-sm">
                                    <span className="w-40 shrink-0 truncate text-slate-600">{pm.model}</span>
                                    <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-200/70">
                                        <div
                                            className="h-full rounded-full bg-slate-400"
                                            style={{ width: `${modelPct}%` }}
                                        />
                                    </div>
                                    <span className="w-16 text-right text-slate-700 tabular-nums">
                                        {fmt(pm.annotated)}
                                    </span>
                                </div>
                            );
                        })
                    )}
                </div>
            ) : null}
        </li>
    );
}

function Summary({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
    return (
        <div className="rounded-xl border border-slate-100 bg-white px-4 py-3">
            <p className="text-xs font-medium text-slate-500">{label}</p>
            <p
                className={`mt-1 text-2xl font-semibold tracking-tight tabular-nums ${accent ? "text-emerald-600" : "text-slate-900"
                    }`}
            >
                {value}
            </p>
        </div>
    );
}

function FamilyChip({ family }: { family: string }) {
    return (
        <span className="hidden rounded-md border border-slate-200 bg-white px-1.5 py-0.5 text-[11px] font-medium text-slate-500 sm:inline">
            {family}
        </span>
    );
}