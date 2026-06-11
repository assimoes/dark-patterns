"use client";

import { useState } from "react";
import { Users } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    Field,
    SubmitButton,
    SuccessPanel,
    inputClass,
} from "@/components/operations/form";
import { useCreateAnnotator } from "@/hooks/useCreateAnnotator";
import { ApiError } from "@/lib/api";
import type { AddAnnotatorInput, AnnotatorKind } from "@/lib/types";

export default function AnnotatorsPage() {
    const [kind, setKind] = useState<AnnotatorKind>("human");
    const [label, setLabel] = useState("");
    const [family, setFamily] = useState("");
    const [slug, setSlug] = useState("");
    const [name, setName] = useState("");
    const [modalities, setModalities] = useState("");

    const createAnnotator = useCreateAnnotator();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: AddAnnotatorInput = { kind, label: label.trim() };
        if (kind === "llm") {
            body.family = family.trim();
            body.slug = slug.trim();
            body.name = name.trim();
            body.modalities = modalities
                .split(",")
                .map((m) => m.trim())
                .filter(Boolean);
        }
        createAnnotator.mutate(body);
    };

    const errorMessage =
        createAnnotator.error instanceof ApiError
            ? createAnnotator.error.message || "Request failed."
            : createAnnotator.isError
                ? "Could not add the annotator. Please try again."
                : null;

    return (
        <Card
            icon={<Users className="size-5" />}
            iconClass="bg-sky-50 text-sky-600"
            title="Add annotator"
            subtitle="Add a human auditor or an llm panel member."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
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

                <Field label="Label" hint="Required for both kinds.">
                    <input
                        type="text"
                        required
                        value={label}
                        onChange={(e) => setLabel(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                {kind === "llm" ? (
                    <>
                        <Field label="Family">
                            <input
                                type="text"
                                required
                                value={family}
                                onChange={(e) => setFamily(e.target.value)}
                                className={inputClass}
                            />
                        </Field>

                        <Field label="Slug">
                            <input
                                type="text"
                                required
                                value={slug}
                                onChange={(e) => setSlug(e.target.value)}
                                className={inputClass}
                            />
                        </Field>

                        <Field label="Model name">
                            <input
                                type="text"
                                required
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                className={inputClass}
                            />
                        </Field>

                        <Field
                            label="Modalities"
                            hint="Comma-separated, e.g. text, image."
                        >
                            <input
                                type="text"
                                value={modalities}
                                onChange={(e) => setModalities(e.target.value)}
                                className={inputClass}
                            />
                        </Field>
                    </>
                ) : null}

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={createAnnotator.isPending}
                    idleLabel="Add annotator"
                    pendingLabel="Adding…"
                />

                {createAnnotator.data ? (
                    <SuccessPanel title="Annotator created" data={createAnnotator.data} />
                ) : null}
            </form>
        </Card>
    );
}