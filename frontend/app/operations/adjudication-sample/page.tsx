"use client";

import { useState } from "react";
import { Shuffle } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    Field,
    SubmitButton,
    SuccessPanel,
    inputClass,
} from "@/components/operations/form";
import { useDashboard } from "@/hooks/useDashboard";
import { useCreateAdjudicationSample } from "@/hooks/useCreateAdjudicationSample";
import { ApiError } from "@/lib/api";
import type { CreateSampleInput, Run } from "@/lib/types";

// Operator form: draw a stratified adjudication sample from a panel run. The run
// is chosen from the dashboard's runs (only those with an llm panel can be
// sampled); the three per-stratum sizes and an optional seed are typed in. On
// success the backend's draw — its sample/gold ids, the seed it used, and the
// realised per-stratum counts — is shown so the operator can reproduce it.

// A panel run is one that has at least one llm member. Human-only/gold runs have
// no panel votes to sample, so they're excluded from the selector.
const isPanelRun = (run: Run) => run.members.some((m) => m.kind === "llm");

export default function AdjudicationSamplePage() {
    const dashboard = useDashboard();
    const panelRuns = (dashboard.data?.runs ?? []).filter(isPanelRun);

    const [panelRunId, setPanelRunId] = useState("");
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
                ? "Could not draw the sample. Please try again."
                : null;

    // Flatten the nested counts so the SuccessPanel renders one row per stratum.
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
        <Card
            icon={<Shuffle className="size-5" />}
            iconClass="bg-violet-50 text-violet-600"
            title="Create adjudication sample"
            subtitle="Draw a stratified sample from a panel run to adjudicate."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
                <Field
                    label="Panel run"
                    hint={
                        dashboard.isPending
                            ? "Loading runs…"
                            : panelRuns.length === 0
                                ? "No panel runs available yet."
                                : "Only runs with an llm panel can be sampled."
                    }
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
                    <input
                        type="number"
                        min={0}
                        required
                        value={flaggedMajority}
                        onChange={(e) => setFlaggedMajority(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Flagged — split" hint="Reviews where the panel was divided.">
                    <input
                        type="number"
                        min={0}
                        required
                        value={flaggedSplit}
                        onChange={(e) => setFlaggedSplit(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Silent" hint="Reviews no model flagged.">
                    <input
                        type="number"
                        min={0}
                        required
                        value={silent}
                        onChange={(e) => setSilent(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Seed" hint="Optional. Omit to let the backend pick (and echo) one.">
                    <input
                        type="number"
                        value={seed}
                        onChange={(e) => setSeed(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={createSample.isPending}
                    idleLabel="Draw sample"
                    pendingLabel="Drawing…"
                    icon={<Shuffle className="size-4" />}
                />

                {result ? <SuccessPanel title="Sample drawn" data={result} /> : null}
            </form>
        </Card>
    );
}
