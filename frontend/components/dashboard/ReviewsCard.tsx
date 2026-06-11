"use client";

import { useState } from "react";
import { ChevronRight, MessageSquare } from "lucide-react";
import { Card, Badge } from "@/components/ui/Card";
import { fmt } from "@/lib/types";
import type { Game, PopulationCoverage, ReviewStat } from "@/lib/types";

import { useGamePopulationModels } from "@/hooks/useGamePopulationModels";
import { usePopulationPanel } from "@/hooks/usePopulationPanel";
import { useGamePopulations } from "@/hooks/useGamePopulations";

export function ReviewsCard({
    games,
    reviewStats,
}: {
    games: Game[];
    reviewStats: ReviewStat[];
}) {
    // Track which games are expanded. A Set lets several stay open at once.
    const [openGames, setOpenGames] = useState<Set<string>>(new Set());
    // Track which populations are expanded, keyed by `gameId:populationId` so the
    // same population id under two games never collides.
    const [openPops, setOpenPops] = useState<Set<string>>(new Set());

    const toggleGame = (id: string) =>
        setOpenGames((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });

    const togglePop = (key: string) =>
        setOpenPops((prev) => {
            const next = new Set(prev);
            if (next.has(key)) next.delete(key);
            else next.add(key);
            return next;
        });

    const gameById = (id: string): Game | undefined => games.find((g) => g.id === id);

    const totalReviews = reviewStats.reduce((s, r) => s + r.reviews, 0);
    const totalAnnotated = reviewStats.reduce((s, r) => s + r.annotated, 0);
    const totalPct = totalReviews === 0 ? 0 : Math.round((totalAnnotated / totalReviews) * 100);

    return (
        <Card
            icon={<MessageSquare className="size-5" />}
            iconClass="bg-emerald-50 text-emerald-600"
            title="Reviews & annotation"
            subtitle="Coverage per game — drill into populations"
            badge={
                <Badge className="border-emerald-200 bg-emerald-50 text-emerald-700">
                    {totalPct}% annotated
                </Badge>
            }
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
                        isOpen={openGames.has(s.gameId)}
                        onToggle={() => toggleGame(s.gameId)}
                        openPops={openPops}
                        onTogglePop={togglePop}
                    />
                ))}
            </ul>
        </Card>
    );
}

function GameRow({
    stat,
    game,
    isOpen,
    onToggle,
    openPops,
    onTogglePop,
}: {
    stat: ReviewStat;
    game: Game | undefined;
    isOpen: boolean;
    onToggle: () => void;
    openPops: Set<string>;
    onTogglePop: (key: string) => void;
}) {
    const populations = useGamePopulations(stat.gameId, isOpen);

    const color = game?.color ?? "#94a3b8";
    const name = game?.name ?? stat.gameId;
    const pct = stat.reviews === 0 ? 0 : Math.round((stat.annotated / stat.reviews) * 100);

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
                        Populations
                    </p>

                    {populations.isPending ? (
                        <p className="text-xs text-slate-400">Loading populations…</p>
                    ) : populations.isError ? (
                        <p className="text-xs text-rose-500">
                            Couldn&apos;t load populations ·{" "}
                            <button
                                type="button"
                                onClick={() => populations.refetch()}
                                className="font-medium underline underline-offset-2 hover:text-rose-600"
                            >
                                Retry
                            </button>
                        </p>
                    ) : populations.data.length === 0 ? (
                        <p className="text-xs text-slate-400">No populations yet.</p>
                    ) : (
                        <ul className="divide-y divide-slate-200/70 overflow-hidden rounded-lg border border-slate-200/70 bg-white">
                            {populations.data.map((p: any) => {
                                const key = `${stat.gameId}:${p.populationId}`;
                                return (
                                    <PopulationRow
                                        key={key}
                                        gameId={stat.gameId}
                                        pop={p}
                                        color={color}
                                        isOpen={openPops.has(key)}
                                        onToggle={() => onTogglePop(key)}
                                    />
                                );
                            })}
                        </ul>
                    )}
                </div>
            ) : null}
        </li>
    );
}

function PopulationRow({
    gameId,
    pop,
    color,
    isOpen,
    onToggle,
}: {
    gameId: string;
    pop: PopulationCoverage;
    color: string;
    isOpen: boolean;
    onToggle: () => void;
}) {
    const popId = String(pop.populationId);
    const models = useGamePopulationModels(gameId, popId, isOpen);
    const panel = usePopulationPanel(popId, isOpen);

    const pct = pop.reviews === 0 ? 0 : Math.round((pop.annotated / pop.reviews) * 100);

    return (
        <li>
            <button
                type="button"
                onClick={onToggle}
                aria-expanded={isOpen}
                className="flex w-full items-center gap-3 px-3.5 py-2.5 text-left text-sm transition-colors hover:bg-slate-50"
            >
                <ChevronRight
                    className={`size-3.5 shrink-0 text-slate-400 transition-transform duration-200 ${isOpen ? "rotate-90" : ""
                        }`}
                />
                <span className="truncate font-medium text-slate-700">{pop.label}</span>

                <span className="ml-auto hidden text-slate-500 tabular-nums sm:inline">
                    <span className="font-medium text-slate-900">{fmt(pop.annotated)}</span> /{" "}
                    {fmt(pop.reviews)}
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
                <div className="space-y-4 border-t border-slate-200/70 bg-slate-50/60 px-3.5 py-3.5">
                    <div className="space-y-2.5">
                        <p className="text-[11px] font-medium uppercase tracking-wide text-slate-400">
                            Annotated by model
                        </p>

                        {models.isPending ? (
                            <p className="text-xs text-slate-400">Loading model coverage…</p>
                        ) : models.isError ? (
                            <p className="text-xs text-rose-500">
                                Couldn&apos;t load model coverage ·{" "}
                                <button
                                    type="button"
                                    onClick={() => models.refetch()}
                                    className="font-medium underline underline-offset-2 hover:text-rose-600"
                                >
                                    Retry
                                </button>
                            </p>
                        ) : models.data.length === 0 ? (
                            <p className="text-xs text-slate-400">No model annotations yet.</p>
                        ) : (
                            models.data.map((pm) => {
                                const modelPct =
                                    pop.reviews === 0 ? 0 : Math.round((pm.annotated / pop.reviews) * 100);
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

                    <div className="space-y-2.5">
                        <p className="text-[11px] font-medium uppercase tracking-wide text-slate-400">
                            Panel members
                        </p>

                        {panel.isPending ? (
                            <p className="text-xs text-slate-400">Loading panel…</p>
                        ) : panel.isError ? (
                            <p className="text-xs text-rose-500">
                                Couldn&apos;t load panel ·{" "}
                                <button
                                    type="button"
                                    onClick={() => panel.refetch()}
                                    className="font-medium underline underline-offset-2 hover:text-rose-600"
                                >
                                    Retry
                                </button>
                            </p>
                        ) : panel.data.length === 0 ? (
                            <p className="text-xs text-slate-400">No panel members.</p>
                        ) : (
                            <div className="flex flex-wrap gap-1.5">
                                {panel.data.map((m, i) => {
                                    const isHuman = m.kind === "human";
                                    return (
                                        <span
                                            key={`${m.kind}:${m.label}:${i}`}
                                            className="inline-flex items-center gap-1.5 rounded-full border border-slate-200 bg-white px-2.5 py-1 text-xs font-medium text-slate-600"
                                        >
                                            <span
                                                className={`size-1.5 shrink-0 rounded-full ${isHuman ? "bg-sky-500" : "bg-violet-500"
                                                    }`}
                                            />
                                            {m.label}
                                            <span className="text-[10px] uppercase tracking-wide text-slate-400">
                                                {m.kind}
                                            </span>
                                        </span>
                                    );
                                })}
                            </div>
                        )}
                    </div>
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