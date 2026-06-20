"use client";

import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { ErrorPanel, Field, SubmitButton, inputClass } from "@/components/operations/form";
import { useListRuns } from "@/hooks/useListRuns";
import { useCreateComparison } from "@/hooks/useComparisons";
import { ApiError } from "@/lib/api/utils";
import type { RunSummary } from "@/lib/types";

// ComparisonBuilder picks two runs. there are no constraints — the comparison runs over whatever reviews
// and models the two runs share, and shows nothing where they don't overlap. onDone receives the new id.
export function ComparisonBuilder({ onDone }: { onDone: (id: number) => void }) {
    const runs = useListRuns();
    const options = runs.data ?? [];

    const [aId, setAId] = useState("");
    const [bId, setBId] = useState("");
    const [label, setLabel] = useState("");

    const create = useCreateComparison();

    const sameRun = aId !== "" && aId === bId;
    const canSubmit = aId !== "" && bId !== "" && !sameRun && label.trim() !== "";

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!canSubmit) return;
        create.mutate(
            { label: label.trim(), runAId: Number(aId), runBId: Number(bId) },
            { onSuccess: (c) => onDone(c.id) },
        );
    };

    const errorMessage =
        create.error instanceof ApiError ? create.error.message || "Request failed." : create.isError ? "Could not create the comparison." : null;

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <Field label="Label">
                <input className={inputClass} value={label} onChange={(e) => setLabel(e.target.value)} placeholder="e.g. expensive vs cheap panel" />
            </Field>

            <div className="grid grid-cols-[1fr_auto_1fr] items-end gap-2">
                <Field label="Run A">
                    <RunPicker value={aId} onChange={setAId} runs={options} />
                </Field>
                <ArrowRight className="mb-2.5 size-4 text-slate-400" />
                <Field label="Run B">
                    <RunPicker value={bId} onChange={setBId} runs={options} />
                </Field>
            </div>

            <p className="text-xs text-slate-400">
                The two runs are compared on the reviews both annotated, model by model. Reviews or models that
                only one run covers are simply left out.
            </p>

            {sameRun ? <ErrorPanel message="Pick two different runs." /> : null}
            <ErrorPanel message={errorMessage} />

            <SubmitButton pending={create.isPending} idleLabel="Create comparison" pendingLabel="Creating…" />
        </form>
    );
}

function RunPicker({ value, onChange, runs }: { value: string; onChange: (v: string) => void; runs: RunSummary[] }) {
    return (
        <select className={inputClass} value={value} onChange={(e) => onChange(e.target.value)} required>
            <option value="">— select —</option>
            {runs.map((r) => (
                <option key={r.id} value={r.id}>
                    #{r.id} · {r.runType} · pop {r.populationId}
                </option>
            ))}
        </select>
    );
}
