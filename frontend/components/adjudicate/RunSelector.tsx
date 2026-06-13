"use client";

import { useDashboard } from "@/hooks/useDashboard";
import type { Run } from "@/lib/types";

// panel run = has at least one llm member; human-only/gold runs have no panel votes, so they're excluded.
const isPanelRun = (run: Run) => run.members.some((m) => m.kind === "llm");

const accentRing: Record<"violet" | "sky", string> = {
    violet: "focus:border-violet-400 focus:ring-violet-100",
    sky: "focus:border-sky-400 focus:ring-sky-100",
};

// run picker shared by the adjudication + blind screens. `accent` only tints the focus ring.
export function RunSelector({
    runId,
    onChange,
    accent = "violet",
}: {
    runId: string;
    onChange: (id: string) => void;
    accent?: "violet" | "sky";
}) {
    const dashboard = useDashboard();
    const panelRuns = (dashboard.data?.runs ?? []).filter(isPanelRun);
    const disabled = dashboard.isPending || panelRuns.length === 0;

    return (
        <select
            value={runId}
            onChange={(e) => onChange(e.target.value)}
            disabled={disabled}
            className={`rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 shadow-sm transition-colors focus:outline-none focus:ring-2 ${accentRing[accent]} disabled:cursor-not-allowed disabled:opacity-50`}
        >
            <option value="">
                {dashboard.isPending
                    ? "Loading runs…"
                    : panelRuns.length === 0
                        ? "No panel runs"
                        : "— select a panel run —"}
            </option>
            {panelRuns.map((run) => (
                <option key={run.id} value={run.id}>
                    #{run.id} · {run.label}
                </option>
            ))}
        </select>
    );
}
