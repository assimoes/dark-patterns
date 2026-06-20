"use client";

import { useState } from "react";
import { ErrorPanel, Field, SubmitButton, inputClass } from "@/components/operations/form";
import { PopulationSelect } from "@/components/operations/PopulationSelect";
import { PromptSelect } from "@/components/operations/PromptSelect";
import { AnnotatorMultiSelect } from "@/components/operations/AnnotatorMultiSelect";
import { useCreateRun } from "@/hooks/useCreateRun";
import { ApiError } from "@/lib/api/utils";
import type { CreateRunInput, RunType } from "@/lib/types";

// RunForm opens an annotation run over a population. create-only; a run's config is pinned. presetPopulationId
// fixes the population when launched from a population row.
export function RunForm({ presetPopulationId, onDone }: { presetPopulationId?: string; onDone: () => void }) {
    const [populationId, setPopulationId] = useState(presetPopulationId ?? "");
    const [promptId, setPromptId] = useState("");
    const [taxonomyVersion, setTaxonomyVersion] = useState("");
    const [runType, setRunType] = useState<"" | RunType>("");
    const [temperature, setTemperature] = useState("");
    const [annotatorIds, setAnnotatorIds] = useState<number[]>([]);

    const createRun = useCreateRun();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: CreateRunInput = { population_id: Number(populationId), prompt_id: Number(promptId) };
        if (taxonomyVersion !== "") body.taxonomy_version = Number(taxonomyVersion);
        if (runType) body.run_type = runType;
        if (temperature !== "") body.temperature = Number(temperature);
        if (annotatorIds.length) body.annotator_ids = annotatorIds;
        createRun.mutate(body, { onSuccess: onDone });
    };

    const errorMessage = runErrorMessage(createRun.error, createRun.isError);

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <PopulationSelect value={populationId} onChange={setPopulationId} required />
            <PromptSelect value={promptId} onChange={setPromptId} required />

            <Field label="Taxonomy version">
                <input type="number" value={taxonomyVersion} onChange={(e) => setTaxonomyVersion(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Run type">
                <select value={runType} onChange={(e) => setRunType(e.target.value as "" | RunType)} className={inputClass}>
                    <option value="">— default —</option>
                    <option value="llm_panel">llm_panel</option>
                    <option value="gold">gold</option>
                </select>
            </Field>

            <Field label="Temperature">
                <input type="number" step="0.1" value={temperature} onChange={(e) => setTemperature(e.target.value)} className={inputClass} />
            </Field>

            <AnnotatorMultiSelect value={annotatorIds} onChange={setAnnotatorIds} />

            <ErrorPanel message={errorMessage} />

            <SubmitButton pending={createRun.isPending} idleLabel="Create run" pendingLabel="Creating…" />
        </form>
    );
}

// the description gate returns a structured 409 ({error, missing:[{id,name}]}); surface it readably.
function runErrorMessage(error: unknown, isError: boolean): string | null {
    if (error instanceof ApiError) {
        try {
            const body = JSON.parse(error.message) as { error?: string; missing?: { name: string }[] };
            if (body.missing?.length) {
                return `${body.error}: ${body.missing.map((m) => m.name).join(", ")}. Approve their descriptions first.`;
            }
            if (body.error) return body.error;
        } catch {
            // not json; fall through to the raw message
        }
        return error.message || "Request failed.";
    }
    return isError ? "Could not create the run." : null;
}
