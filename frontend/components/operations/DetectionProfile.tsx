"use client";

import { useMemo } from "react";
import { codebook, families } from "@/lib/codebook";
import type { PatternCount } from "@/lib/types";

const nameByCode = new Map(codebook.map((p) => [p.code, p.name]));
const familyByCode = new Map(codebook.map((p) => [p.code, p.family]));
const familyOrder = Object.keys(families) as (keyof typeof families)[];

// DetectionProfile shows a run's most-flagged patterns and family mix for a scope (one game or all
// games). counts are reviews where the run's panel majority flagged the code.
export function DetectionProfile({ rows, gameId }: { rows: PatternCount[]; gameId: string }) {
    const { byFamily, topCodes, total, maxFamily } = useMemo(() => {
        const scoped = gameId === "all" ? rows : rows.filter((r) => r.gameId === gameId);
        const codeTotals = new Map<string, number>();
        for (const r of scoped) codeTotals.set(r.code, (codeTotals.get(r.code) ?? 0) + r.reviews);

        const fam = new Map<string, number>();
        for (const [code, n] of codeTotals) {
            const f = familyByCode.get(code);
            if (f) fam.set(f, (fam.get(f) ?? 0) + n);
        }
        const top = [...codeTotals.entries()].sort((a, b) => b[1] - a[1]).slice(0, 8);
        const sum = [...codeTotals.values()].reduce((s, n) => s + n, 0);
        return { byFamily: fam, topCodes: top, total: sum, maxFamily: Math.max(1, ...fam.values()) };
    }, [rows, gameId]);

    if (total === 0) return <p className="py-2 text-sm text-slate-400">No detections in this scope.</p>;

    const topFamily = [...byFamily.entries()].sort((a, b) => b[1] - a[1])[0];

    return (
        <div className="flex flex-col gap-4">
            <div>
                <h5 className="mb-1.5 text-[11px] font-semibold text-slate-400">
                    Family mix {topFamily ? <span className="font-normal text-slate-300">· most: {topFamily[0]}</span> : null}
                </h5>
                <div className="flex flex-col gap-1">
                    {familyOrder.map((f) => {
                        const n = byFamily.get(f) ?? 0;
                        const pct = Math.round((n / maxFamily) * 100);
                        return (
                            <div key={f} className="flex items-center gap-2 text-xs">
                                <span className="w-9 shrink-0 font-medium text-slate-500" title={families[f]}>{f}</span>
                                <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                                    <div className="h-full rounded-full bg-violet-400" style={{ width: `${pct}%` }} />
                                </div>
                                <span className="w-8 shrink-0 text-right tabular-nums text-slate-500">{n}</span>
                            </div>
                        );
                    })}
                </div>
            </div>

            <div>
                <h5 className="mb-1.5 text-[11px] font-semibold text-slate-400">Top patterns</h5>
                <div className="flex flex-col gap-1">
                    {topCodes.map(([code, n]) => (
                        <div key={code} className="flex items-center justify-between gap-2 text-xs">
                            <span className="flex min-w-0 items-center gap-2">
                                <span className="font-medium text-slate-700">{code}</span>
                                <span className="truncate text-slate-400">{nameByCode.get(code) ?? ""}</span>
                            </span>
                            <span className="shrink-0 tabular-nums text-slate-500">{n}</span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
