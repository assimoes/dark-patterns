"use client";

import { Field, inputClass } from "@/components/operations/form";
import { useListPrompts } from "@/hooks/useListPrompts";

export function PromptSelect({
    value,
    onChange,
    required = false,
}: {
    value: string;
    onChange: (value: string) => void;
    required?: boolean;
}) {
    const { data, isPending, isError } = useListPrompts();
    const prompts = data ?? [];

    const hint = isPending
        ? "Loading prompts…"
        : isError
            ? "Couldn't load prompts."
            : prompts.length === 0
                ? "No prompts available."
                : "Pick the prompt the run annotates with.";

    return (
        <Field label="Prompt" hint={hint}>
            <select
                required={required}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className={inputClass}
                disabled={isPending || isError || prompts.length === 0}
            >
                <option value="">— select a prompt —</option>
                {prompts.map((p) => (
                    <option key={p.id} value={p.id}>
                        #{p.id} · {p.name} v{p.version} ({p.modality})
                    </option>
                ))}
            </select>
        </Field>
    );
}