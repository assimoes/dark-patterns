"use client";

import { useState } from "react";
import { ErrorPanel, Field, SubmitButton, inputClass } from "@/components/operations/form";
import { useCreateAnnotator } from "@/hooks/useCreateAnnotator";
import { useUpdateAnnotator } from "@/hooks/useEntityMutations";
import { ApiError } from "@/lib/api/utils";
import type { AddAnnotatorInput, AnnotatorKind, AnnotatorRow } from "@/lib/types";

// AnnotatorForm creates a human/llm annotator, or edits the label when `editing` is set (label is the
// only mutable field).
export function AnnotatorForm({ editing, onDone }: { editing?: AnnotatorRow; onDone: () => void }) {
    const isEdit = Boolean(editing);
    const [kind, setKind] = useState<AnnotatorKind>((editing?.kind as AnnotatorKind) ?? "human");
    const [label, setLabel] = useState(editing?.label ?? "");
    const [family, setFamily] = useState("");
    const [slug, setSlug] = useState("");
    const [name, setName] = useState("");
    const [modalities, setModalities] = useState("");

    const create = useCreateAnnotator();
    const update = useUpdateAnnotator();
    const mut = isEdit ? update : create;

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (isEdit) {
            update.mutate({ id: String(editing!.id), body: { label: label.trim() } }, { onSuccess: onDone });
            return;
        }
        const body: AddAnnotatorInput = { kind, label: label.trim() };
        if (kind === "llm") {
            body.family = family.trim();
            body.slug = slug.trim();
            body.name = name.trim();
            body.modalities = modalities.split(",").map((m) => m.trim()).filter(Boolean);
        }
        create.mutate(body, { onSuccess: onDone });
    };

    const errorMessage =
        mut.error instanceof ApiError ? mut.error.message || "Request failed." : mut.isError ? "Request failed." : null;

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            {!isEdit ? (
                <Field label="Kind">
                    <div className="inline-flex rounded-lg border border-slate-200 bg-slate-50 p-0.5">
                        {(["human", "llm"] as const).map((k) => (
                            <button
                                key={k}
                                type="button"
                                onClick={() => setKind(k)}
                                className={
                                    kind === k
                                        ? "rounded-md bg-slate-900 px-4 py-1.5 text-sm font-medium text-white transition-colors"
                                        : "rounded-md px-4 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-slate-900"
                                }
                            >
                                {k}
                            </button>
                        ))}
                    </div>
                </Field>
            ) : null}

            <Field label="Label" hint="Required.">
                <input type="text" required value={label} onChange={(e) => setLabel(e.target.value)} className={inputClass} />
            </Field>

            {!isEdit && kind === "llm" ? (
                <>
                    <Field label="Family">
                        <input type="text" required value={family} onChange={(e) => setFamily(e.target.value)} className={inputClass} />
                    </Field>
                    <Field label="Slug">
                        <input type="text" required value={slug} onChange={(e) => setSlug(e.target.value)} className={inputClass} />
                    </Field>
                    <Field label="Model name">
                        <input type="text" required value={name} onChange={(e) => setName(e.target.value)} className={inputClass} />
                    </Field>
                    <Field label="Modalities" hint="Comma-separated, e.g. text, image.">
                        <input type="text" value={modalities} onChange={(e) => setModalities(e.target.value)} className={inputClass} />
                    </Field>
                </>
            ) : null}

            <ErrorPanel message={errorMessage} />

            <SubmitButton
                pending={mut.isPending}
                idleLabel={isEdit ? "Save label" : "Add annotator"}
                pendingLabel={isEdit ? "Saving…" : "Adding…"}
            />
        </form>
    );
}
