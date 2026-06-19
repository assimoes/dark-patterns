"use client";

import { useState } from "react";
import { Layers } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    Field,
    SubmitButton,
    SuccessPanel,
    inputClass,
} from "@/components/operations/form";
import { useCreatePopulation } from "@/hooks/useCreatePopulation";
import { useDashboard } from "@/hooks/useDashboard";
import { ApiError } from "@/lib/api/utils";
import type { CreatePopulationInput } from "@/lib/types";

export default function PopulationsPage() {
    const [description, setDescription] = useState("");
    const [minHours, setMinHours] = useState("");
    const [perGameCap, setPerGameCap] = useState("");
    const [cutoff, setCutoff] = useState("");
    // the external ids of the games to include. empty means every game, the backend default.
    const [games, setGames] = useState<Set<string>>(new Set());
    const [modality, setModality] = useState<"text" | "image">("text");

    const createPopulation = useCreatePopulation();

    // the game list to pick from comes from the dashboard, already cached for the rest of the app.
    const dashboard = useDashboard();
    const allGames = dashboard.data?.games ?? [];

    const toggleGame = (id: string) =>
        setGames((prev) => {
            const next = new Set(prev);
            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }
            return next;
        });

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: CreatePopulationInput = {};
        if (description.trim()) body.description = description.trim();
        if (minHours !== "") body.min_hours_played = Number(minHours);
        if (perGameCap !== "") body.per_game_cap = Number(perGameCap);
        // datetime-local has no tz; toISOString() gives the RFC3339 UTC the backend wants.
        if (cutoff) body.artifacts_cutoff = new Date(cutoff).toISOString();
        // only send game_ids when some are picked, an empty list would mean "use every game" anyway.
        if (games.size > 0) body.game_ids = [...games].map(Number);
        if (modality === 'image') body.modality = 'image'
        createPopulation.mutate(body);
    };

    const errorMessage =
        createPopulation.error instanceof ApiError
            ? createPopulation.error.message || "Request failed."
            : createPopulation.isError
                ? "Could not create the population. Please try again."
                : null;

    return (
        <Card
            icon={<Layers className="size-5" />}
            iconClass="bg-violet-50 text-violet-600"
            title="Create population"
            subtitle="Materialise a population from filter criteria."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
                <Field label="Description">
                    <input
                        type="text"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Min hours played">
                    <input
                        type="number"
                        min="0"
                        value={minHours}
                        onChange={(e) => setMinHours(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Per-game cap">
                    <input
                        type="number"
                        min="0"
                        value={perGameCap}
                        onChange={(e) => setPerGameCap(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field
                    label="Games"
                    hint={
                        games.size === 0
                            ? "None picked means every game is included."
                            : `${games.size} picked. Only these games will be in the population.`
                    }
                >
                    {allGames.length === 0 ? (
                        <p className="text-sm text-slate-400">No games loaded yet.</p>
                    ) : (
                        <div className="flex flex-wrap gap-2">
                            {allGames.map((g) => {
                                const on = games.has(g.id);
                                return (
                                    <button
                                        key={g.id}
                                        type="button"
                                        onClick={() => toggleGame(g.id)}
                                        aria-pressed={on}
                                        className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm transition-colors ${on
                                            ? "border-violet-300 bg-violet-50 text-violet-700"
                                            : "border-slate-200 bg-white text-slate-600 hover:bg-slate-50"
                                            }`}
                                    >
                                        <span className="size-2 rounded-full" style={{ backgroundColor: g.color }} />
                                        {g.name}
                                    </button>
                                );
                            })}
                        </div>
                    )}
                </Field>
                <Field label="Modality" hint="Which artifact kind to freeze.">
                    <div className="inline-flex rounded-lg border border-slate-200 bg-white p-0.5">
                        {(["text", "image"] as const).map((m) => (
                            <button
                                key={m}
                                type="button"
                                onClick={() => setModality(m)}
                                aria-pressed={modality === m}
                                className={`rounded-md px-3 py-1.5 text-sm font-medium capitalize transition-colors ${modality === m
                                    ? "bg-violet-50 text-violet-700"
                                    : "text-slate-500 hover:text-slate-700"
                                    }`}
                            >
                                {m}
                            </button>
                        ))}
                    </div>
                </Field>
                <Field
                    label="Artifacts cutoff"
                    hint="Only include reviews before this moment."
                >
                    <input
                        type="datetime-local"
                        value={cutoff}
                        onChange={(e) => setCutoff(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={createPopulation.isPending}
                    idleLabel="Create population"
                    pendingLabel="Creating…"
                />

                {createPopulation.data ? (
                    <SuccessPanel
                        title="Population created"
                        data={createPopulation.data}
                    />
                ) : null}
            </form>
        </Card>
    );
}