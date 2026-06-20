"use client";

import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ErrorPanel, Field, SubmitButton, inputClass } from "@/components/operations/form";
import { useCreatePrompt, useUpdatePrompt } from "@/hooks/useEntityMutations";
import { api } from "@/lib/api";
import { ApiError } from "@/lib/api/utils";
import type { PromptInput } from "@/lib/types";

// PromptForm creates a prompt, or edits one when `editingId` is set. an edit fetches the full body
// (template + system prompt) to prefill, since the list row does not carry it.
export function PromptForm({ editingId, onDone }: { editingId?: number; onDone: () => void }) {
    const isEdit = editingId != null;

    const detail = useQuery({
        queryKey: ["prompt", editingId],
        queryFn: () => api.browse.getPrompt(editingId!),
        enabled: isEdit,
    });

    const [name, setName] = useState("");
    const [version, setVersion] = useState("");
    const [modality, setModality] = useState("text");
    const [systemPrompt, setSystemPrompt] = useState("");
    const [template, setTemplate] = useState("");

    // prefill once the edit detail loads.
    useEffect(() => {
        if (detail.data) {
            setName(detail.data.name);
            setVersion(String(detail.data.version));
            setModality(detail.data.modality);
            setSystemPrompt(detail.data.system_prompt);
            setTemplate(detail.data.template);
        }
    }, [detail.data]);

    const create = useCreatePrompt();
    const update = useUpdatePrompt();
    const mut = isEdit ? update : create;

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: PromptInput = {
            name: name.trim(),
            version: Number(version),
            modality,
            template,
        };
        if (systemPrompt.trim()) body.system_prompt = systemPrompt;
        if (isEdit) {
            update.mutate({ id: String(editingId), body }, { onSuccess: onDone });
        } else {
            create.mutate(body, { onSuccess: onDone });
        }
    };

    const errorMessage =
        mut.error instanceof ApiError ? mut.error.message || "Request failed." : mut.isError ? "Request failed." : null;

    if (isEdit && detail.isPending) {
        return <p className="text-sm text-slate-400">Loading prompt…</p>;
    }

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <Field label="Name">
                <input type="text" required value={name} onChange={(e) => setName(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Version">
                <input type="number" required min="1" value={version} onChange={(e) => setVersion(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Modality">
                <select value={modality} onChange={(e) => setModality(e.target.value)} className={inputClass}>
                    <option value="text">text</option>
                    <option value="image">image</option>
                    <option value="multimodal">multimodal</option>
                </select>
            </Field>

            <Field label="System prompt" hint="Optional.">
                <textarea
                    rows={4}
                    value={systemPrompt}
                    onChange={(e) => setSystemPrompt(e.target.value)}
                    className={`${inputClass} font-mono`}
                />
            </Field>

            <Field label="Template" hint="The per-item user template. Required.">
                <textarea
                    rows={6}
                    required
                    value={template}
                    onChange={(e) => setTemplate(e.target.value)}
                    className={`${inputClass} font-mono`}
                />
            </Field>

            <ErrorPanel message={errorMessage} />

            <SubmitButton
                pending={mut.isPending}
                idleLabel={isEdit ? "Save prompt" : "Add prompt"}
                pendingLabel={isEdit ? "Saving…" : "Adding…"}
            />
        </form>
    );
}
