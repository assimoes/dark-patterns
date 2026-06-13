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
import { ApiError } from "@/lib/api/utils";
import type { CreatePopulationInput } from "@/lib/types";

export default function PopulationsPage() {
    const [description, setDescription] = useState("");
    const [minHours, setMinHours] = useState("");
    const [perGameCap, setPerGameCap] = useState("");
    const [cutoff, setCutoff] = useState("");

    const createPopulation = useCreatePopulation();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: CreatePopulationInput = {};
        if (description.trim()) body.description = description.trim();
        if (minHours !== "") body.min_hours_played = Number(minHours);
        if (perGameCap !== "") body.per_game_cap = Number(perGameCap);
        // datetime-local has no tz; toISOString() gives the RFC3339 UTC the backend wants.
        if (cutoff) body.artifacts_cutoff = new Date(cutoff).toISOString();
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