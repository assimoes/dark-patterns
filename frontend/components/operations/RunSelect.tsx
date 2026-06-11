"use client";

import { Field, inputClass } from "@/components/operations/form";
import { useListRuns } from "@/hooks/useListRuns";


export function RunSelect({
    value,
    onChange,
    runTypeFilter,
    label = "Run",
    required = false,
}: {
    value: string;
    onChange: (value: string) => void;
    runTypeFilter?: string;
    label?: string;
    required?: boolean;
}) {
    const { data, isPending, isError } = useListRuns();
    const runs = (data ?? []).filter(
        (r) => !runTypeFilter || r.runType === runTypeFilter,
    );

    const hint = isPending
        ? "Loading runs…"
        : isError
            ? "Couldn't load runs."
            : runs.length === 0
                ? runTypeFilter
                    ? `No ${runTypeFilter} runs available yet.`
                    : "No runs available yet."
                : "Pick the run.";

    return (
        <Field label={label} hint={hint}>
            <select
                required={required}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className={inputClass}
                disabled={isPending || isError || runs.length === 0}
            >
                <option value="">— select a run —</option>
                {runs.map((r) => (
                    <option key={r.id} value={r.id}>
                        #{r.id} · {r.label} ({r.population})
                    </option>
                ))}
            </select>
        </Field>
    );
}