"use client";

import { useState } from "react";
import { Plus, X } from "lucide-react";
import { ErrorPanel, Field, SubmitButton, inputClass } from "@/components/operations/form";
import { useCreateGame } from "@/hooks/useCreateGame";
import { useUpdateGame } from "@/hooks/useEntityMutations";
import { ApiError } from "@/lib/api/utils";
import type { CreateGameInput, GameRow, Monetization, UpdateGameInput } from "@/lib/types";

type Ref = { key: string; value: string };

// GameForm creates a game, or edits one when `editing` is set. the id is auto-assigned by the backend.
// source refs are a free key/value list (steam -> app id, reddit -> subreddit, any future source) so a
// new source needs no form change.
export function GameForm({ editing, onDone }: { editing?: GameRow; onDone: () => void }) {
    const isEdit = Boolean(editing);
    const [name, setName] = useState(editing?.name ?? "");
    const [short, setShort] = useState(editing?.short ?? "");
    const [monetization, setMonetization] = useState<"" | Monetization>(
        (editing?.monetization as Monetization) ?? "",
    );
    const [color, setColor] = useState(editing?.color ?? "#6366f1");
    const [refs, setRefs] = useState<Ref[]>(
        editing ? Object.entries(editing.sourceRefs).map(([key, value]) => ({ key, value })) : [],
    );

    const create = useCreateGame();
    const update = useUpdateGame();
    const mut = isEdit ? update : create;

    const setRef = (i: number, patch: Partial<Ref>) =>
        setRefs((prev) => prev.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));
    const addRef = () => setRefs((prev) => [...prev, { key: "", value: "" }]);
    const removeRef = (i: number) => setRefs((prev) => prev.filter((_, idx) => idx !== i));

    // collapse the key/value rows into a map, dropping blank keys.
    const refsObject = () => {
        const out: Record<string, string> = {};
        for (const r of refs) {
            const k = r.key.trim();
            if (k) out[k] = r.value.trim();
        }
        return out;
    };

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (isEdit) {
            const body: UpdateGameInput = {
                name: name.trim(),
                short: short.trim(),
                monetization: (monetization || "f2p") as Monetization,
                color,
                source_refs: refsObject(),
            };
            update.mutate({ id: editing!.id, body }, { onSuccess: onDone });
        } else {
            const body: CreateGameInput = { name: name.trim() };
            if (short.trim()) body.short = short.trim();
            if (monetization) body.monetization = monetization;
            if (color) body.color = color;
            const refsObj = refsObject();
            if (Object.keys(refsObj).length > 0) body.source_refs = refsObj;
            create.mutate(body, { onSuccess: onDone });
        }
    };

    const errorMessage =
        mut.error instanceof ApiError ? mut.error.message || "Request failed." : mut.isError ? "Request failed." : null;

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <Field label="Name" hint="Required.">
                <input type="text" required value={name} onChange={(e) => setName(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Short label">
                <input type="text" value={short} onChange={(e) => setShort(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Monetization">
                <select
                    value={monetization}
                    onChange={(e) => setMonetization(e.target.value as "" | Monetization)}
                    className={inputClass}
                >
                    <option value="">— default —</option>
                    <option value="f2p">f2p</option>
                    <option value="premium">premium</option>
                </select>
            </Field>

            <Field label="Accent color">
                <input
                    type="color"
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    className="h-10 w-16 cursor-pointer rounded-lg border border-slate-200 bg-white p-1"
                />
            </Field>

            <Field
                label="Source references"
                hint="Where each source scrapes from: a source key (steam, reddit, …) and its handle (app id, subreddit)."
            >
                <div className="flex flex-col gap-2">
                    {refs.map((r, i) => (
                        <div key={i} className="flex items-center gap-2">
                            <input
                                type="text"
                                placeholder="source"
                                value={r.key}
                                onChange={(e) => setRef(i, { key: e.target.value })}
                                className={`${inputClass} w-28`}
                            />
                            <input
                                type="text"
                                placeholder="handle"
                                value={r.value}
                                onChange={(e) => setRef(i, { value: e.target.value })}
                                className={`${inputClass} flex-1`}
                            />
                            <button
                                type="button"
                                onClick={() => removeRef(i)}
                                className="grid size-8 shrink-0 place-items-center rounded-lg text-slate-400 transition-colors hover:bg-slate-50 hover:text-slate-700"
                                aria-label="Remove"
                            >
                                <X className="size-4" />
                            </button>
                        </div>
                    ))}
                    <button
                        type="button"
                        onClick={addRef}
                        className="inline-flex w-fit items-center gap-1.5 rounded-lg border border-slate-200 px-2.5 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50"
                    >
                        <Plus className="size-3.5" /> Add source
                    </button>
                </div>
            </Field>

            <ErrorPanel message={errorMessage} />

            <SubmitButton
                pending={mut.isPending}
                idleLabel={isEdit ? "Save game" : "Add game"}
                pendingLabel={isEdit ? "Saving…" : "Adding…"}
            />
        </form>
    );
}
