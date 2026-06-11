"use client";

import { Field, inputClass } from "@/components/operations/form";
import { useListGames } from "@/hooks/useListGames";

export function GameSelect({
    value,
    onChange,
    required = false,
}: {
    value: string;
    onChange: (value: string) => void;
    required?: boolean;
}) {
    const { data, isPending, isError } = useListGames();
    const prompts = data ?? [];

    const hint = isPending
        ? "Loading games…"
        : isError
            ? "Couldn't load games."
            : prompts.length === 0
                ? "No games available."
                : "Pick the game to scrape from steam.";

    return (
        <Field label="Game" hint={hint}>
            <select
                required={required}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className={inputClass}
                disabled={isPending || isError || prompts.length === 0}
            >
                <option value="">— select a game —</option>
                {prompts.map((p) => (
                    <option key={p.id} value={p.id}>
                        #{p.id} · {p.name}
                    </option>
                ))}
            </select>
        </Field>
    );
}