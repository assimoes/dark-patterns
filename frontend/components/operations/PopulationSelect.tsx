"use client";

import { Field, inputClass } from "@/components/operations/form";
import { useListPopulations } from "@/hooks/useListPopulations";


export function PopulationSelect({
    value,
    onChange,
    required = false,
}: {
    value: string;
    onChange: (value: string) => void;
    required?: boolean;
}) {
    const { data, isPending, isError } = useListPopulations();
    const populations = data ?? [];

    const hint = isPending
        ? "Loading populations…"
        : isError
            ? "Couldn't load populations."
            : populations.length === 0
                ? "No populations yet — create one first."
                : "Pick the population the run draws from.";

    return (
        <Field label="Population" hint={hint}>
            <select
                required={required}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className={inputClass}
                disabled={isPending || isError || populations.length === 0}
            >
                <option value="">— select a population —</option>
                {populations.map((p) => (
                    <option key={p.id} value={p.id}>
                        #{p.id} · {p.label} ({p.modality})
                    </option>
                ))}
            </select>
        </Field>
    );
}