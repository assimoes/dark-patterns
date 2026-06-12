"use client";

import { useMemo, useState } from "react";
import { ChevronRight, MessageSquare } from "lucide-react";
import { Card, Badge } from "@/components/ui/Card";
import { fmt } from "@/lib/types";
import type { Game, ReviewStat } from "@/lib/types";

import { useGameRunModels } from "@/hooks/useGameRunModels";

export function ReviewsCard({
    games,
    reviewStats,
}: {
    games: Game[];
    reviewStats: ReviewStat[];
}) {
    const [openGames, setOpenGames] = useState<Set<string>>(new Set());
    const [openRuns, setOpenRuns] = useState<Set<string>>(new Set());

    const toggleGame = (id: string) =>
        setOpenGames((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });

    const toggleRun = (key: string) =>
        setOpenRuns((prev) => {
            const next = new Set(prev);
            if (next.has(key)) next.delete(key);
            else next.add(key);
            return next;
        });

    const statsByGame = useMemo(() => {
        const map = new Map<string, ReviewStat[]>();
        for (const s of reviewStats) {
            const list = map.get(s.gameId);
            if (list) list.push(s);
            else map.set(s.gameId, [s]);
        }
        return map;
    }, [reviewStats]);

    const gameById = (id: string): Game | undefined => games.find((g) => g.id === id);

    const totalReviews = reviewStats.reduce((s, r) => s + r.reviews, 0);
    const totalAnnotated = reviewStats.reduce((s, r) => s + r.annotated, 0);
    const totalPct = totalReviews === 0 ? 0 : Math.round((totalAnnotated / totalReviews) * 100);

    return (
        <Card
            icon={<MessageSquare className="size-5" />}
            iconClass="bg-emerald-50 text-emerald-600"
            title="Reviews & annotation"
            subtitle="Coverage per game — drill into runs"
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
                {Array.from(statsByGame.entries()).map(([gameId, runs]) => (
                    <GameRow
                        key={gameId}
                        gameId={gameId}
                        runs={runs}
                        game={gameById(gameId)}
                        isOpen={openGames.has(gameId)}
                        onToggle={() => toggleGame(gameId)}
                        openRuns={openRuns}
                        onToggleRun={toggleRun}
                    />
                ))}
            </ul>
        </Card>
    );
}

function GameRow({
    gameId,
    runs,
    game,
    isOpen,
    onToggle,
    openRuns,
    onToggleRun,
}: {
    gameId: string;
    runs: ReviewStat[];
    game: Game | undefined;
    isOpen: boolean;
    onToggle: () => void;
    openRuns: Set<string>;
    onToggleRun: (key: string) => void;
}) {
    const color = game?.color ?? "#94a3b8";
    const name = game?.name ?? gameId;

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

                <span className="ml-auto text-xs font-medium text-slate-500 tabular-nums">
                    {runs.length} {runs.length === 1 ? "run" : "runs"}
                </span>
            </button>

            {isOpen ? (
                <div className="space-y-2.5 border-t border-slate-100 bg-slate-50/60 px-4 py-3.5">
                    <p className="text-[11px] font-medium uppercase tracking-wide text-slate-400">
                        Runs
                    </p>

                    {runs.length === 0 ? (
                        <p className="text-xs text-slate-400">No runs yet.</p>
                    ) : (
                        <ul className="divide-y divide-slate-200/70 overflow-hidden rounded-lg border border-slate-200/70 bg-white">
                            {runs.map((run) => {
                                const key = `${gameId}:${run.runId}`;
                                return (
                                    <RunRow
                                        key={key}
                                        gameId={gameId}
                                        run={run}
                                        color={color}
                                        isOpen={openRuns.has(key)}
                                        onToggle={() => onToggleRun(key)}
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

function RunRow({
    gameId,
    run,
    color,
    isOpen,
    onToggle,
}: {
    gameId: string;
    run: ReviewStat;
    color: string;
    isOpen: boolean;
    onToggle: () => void;
}) {
    const models = useGameRunModels(gameId, run.runId, isOpen);

    const pct = run.reviews === 0 ? 0 : Math.round((run.annotated / run.reviews) * 100);

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
                <span className="truncate font-medium text-slate-700">{run.prompt}</span>
                <span className="shrink-0 text-xs text-slate-400 tabular-nums">#{run.runId}</span>

                <span className="ml-auto hidden text-slate-500 tabular-nums sm:inline">
                    <span className="font-medium text-slate-900">{fmt(run.annotated)}</span> /{" "}
                    {fmt(run.reviews)}
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
                <div className="space-y-2.5 border-t border-slate-200/70 bg-slate-50/60 px-3.5 py-3.5">
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
                                run.reviews === 0 ? 0 : Math.round((pm.annotated / run.reviews) * 100);
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
