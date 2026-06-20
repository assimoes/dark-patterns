"use client";

import { useState } from "react";
import { Shuffle } from "lucide-react";
import { ErrorPanel, Field, SubmitButton, SuccessPanel, inputClass } from "@/components/operations/form";
import { useDashboard } from "@/hooks/useDashboard";
import { useCreateAdjudicationSample } from "@/hooks/useCreateAdjudicationSample";
import { ApiError } from "@/lib/api/utils";
import type { CreateSampleInput, Run } from "@/lib/types";

const isPanelRun = (run: Run) => run.members.some((m) => m.kind === "llm");

// AdjSampleForm draws a stratified adjudication sample from a panel run. presetPanelRunId fixes the run
// when launched from a run row.
export function AdjSampleForm({ presetPanelRunId }: { presetPanelRunId?: string }) {
    const dashboard = useDashboard();
    const panelRuns = (dashboard.data?.runs ?? []).filter(isPanelRun);

    const [panelRunId, setPanelRunId] = useState(presetPanelRunId ?? "");
    const [flaggedMajority, setFlaggedMajority] = useState("");
    const [flaggedSplit, setFlaggedSplit] = useState("");
    const [silent, setSilent] = useState("");
    const [seed, setSeed] = useState("");

    const createSample = useCreateAdjudicationSample();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: CreateSampleInput = {
            panelRunId: Number(panelRunId),
            flaggedMajority: Number(flaggedMajority),
            flaggedSplit: Number(flaggedSplit),
            silent: Number(silent),
        };
        if (seed !== "") body.seed = Number(seed);
        createSample.mutate(body);
    };

    const errorMessage =
        createSample.error instanceof ApiError
            ? createSample.error.message || "Request failed."
            : createSample.isError
                ? "Could not draw the sample."
                : null;

    const result = createSample.data
        ? {
            sampleId: createSample.data.sampleId,
            goldRunId: createSample.data.goldRunId,
            seed: createSample.data.seed,
            flagged_majority: createSample.data.counts.flagged_majority,
            flagged_split: createSample.data.counts.flagged_split,
            silent: createSample.data.counts.silent,
        }
        : null;

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <Field
                label="Panel run"
                hint={panelRuns.length === 0 ? "No panel runs available yet." : "Only runs with an llm panel can be sampled."}
            >
                <select
                    required
                    value={panelRunId}
                    onChange={(e) => setPanelRunId(e.target.value)}
                    className={inputClass}
                    disabled={panelRuns.length === 0}
                >
                    <option value="">— select a run —</option>
                    {panelRuns.map((run) => (
                        <option key={run.id} value={run.id}>
                            #{run.id} · {run.label}
                        </option>
                    ))}
                </select>
            </Field>

            <Field label="Flagged — majority" hint="Reviews where the panel agreed.">
                <input type="number" min={0} required value={flaggedMajority} onChange={(e) => setFlaggedMajority(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Flagged — split" hint="Reviews where the panel was divided.">
                <input type="number" min={0} required value={flaggedSplit} onChange={(e) => setFlaggedSplit(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Silent" hint="Reviews no model flagged.">
                <input type="number" min={0} required value={silent} onChange={(e) => setSilent(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Seed" hint="Optional. Omit to let the backend pick one.">
                <input type="number" value={seed} onChange={(e) => setSeed(e.target.value)} className={inputClass} />
            </Field>

            <ErrorPanel message={errorMessage} />

            <SubmitButton pending={createSample.isPending} idleLabel="Draw sample" pendingLabel="Drawing…" icon={<Shuffle className="size-4" />} />

            {result ? <SuccessPanel title="Sample drawn" data={result} /> : null}
        </form>
    );
}
