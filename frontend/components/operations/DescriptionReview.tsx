"use client";

import { useState } from "react";
import { AlertTriangle, Check, Plus, RefreshCw, Save, X } from "lucide-react";
import { Field, inputClass } from "@/components/operations/form";
import {
    useApproveDescription,
    useGameDescriptions,
    useResearchGame,
    useUpdateDescription,
} from "@/hooks/useGameDescriptions";
import { ApiError } from "@/lib/api/utils";
import { emptyProfile } from "@/lib/types";
import type {
    Acquisition,
    DescriptionProfile,
    GameDescription,
    Resource,
} from "@/lib/types";

// DescriptionReview is the human gate: review the researched draft, fix it, and approve it as the frozen
// version. an approved description is read-only; re-research opens a fresh draft.
export function DescriptionReview({ gameId, gameName }: { gameId: string; gameName: string }) {
    const { data, isPending, isError } = useGameDescriptions(gameId);
    const research = useResearchGame(gameId);

    if (isPending) return <p className="text-sm text-slate-400">Loading description…</p>;
    if (isError) return <p className="text-sm text-rose-600">Could not load the description.</p>;

    const versions = data ?? [];
    const active = versions[0];
    const approved = versions.find((v) => v.status === "approved");

    if (!active) {
        return (
            <div className="flex flex-col gap-3">
                <p className="text-sm text-slate-500">
                    No description yet for <span className="font-medium text-slate-700">{gameName}</span>.
                </p>
                <ResearchButton
                    label="Research now"
                    pending={research.isPending}
                    done={research.isSuccess}
                    onClick={() => research.mutate(undefined)}
                />
            </div>
        );
    }

    // an approved version with no newer draft: read-only, with re-research to start a new one.
    if (active.status === "approved") {
        return (
            <div className="flex flex-col gap-4">
                <StatusBadge status="approved" />
                <RenderedPreview text={active.renderedText} />
                <p className="text-xs text-slate-400">
                    Approved v{active.version}. Editing requires a new version — re-research to draft one.
                </p>
                <ResearchButton
                    label="Re-research"
                    pending={research.isPending}
                    done={research.isSuccess}
                    onClick={() => research.mutate(undefined)}
                />
            </div>
        );
    }

    return (
        <DraftEditor
            key={active.id}
            gameId={gameId}
            draft={active}
            approvedVersion={approved?.version}
            researchPending={research.isPending}
            researchDone={research.isSuccess}
            onResearch={() => research.mutate(undefined)}
        />
    );
}

function DraftEditor({
    gameId,
    draft,
    approvedVersion,
    researchPending,
    researchDone,
    onResearch,
}: {
    gameId: string;
    draft: GameDescription;
    approvedVersion?: number;
    researchPending: boolean;
    researchDone: boolean;
    onResearch: () => void;
}) {
    const [profile, setProfile] = useState<DescriptionProfile>({ ...emptyProfile, ...draft.profile });
    const [dirty, setDirty] = useState(false);

    const update = useUpdateDescription(gameId);
    const approve = useApproveDescription(gameId);

    const edit = (patch: Partial<DescriptionProfile>) => {
        setProfile((p) => ({ ...p, ...patch }));
        setDirty(true);
    };
    const editBM = (patch: Partial<DescriptionProfile["business_model"]>) =>
        edit({ business_model: { ...profile.business_model, ...patch } });

    const save = () =>
        update.mutate(
            { id: draft.id, profile },
            { onSuccess: () => setDirty(false) },
        );

    const flags = draft.valenceFlags ?? [];
    const canApprove = !dirty && flags.length === 0;

    const errorMessage = (m: typeof update | typeof approve) =>
        m.error instanceof ApiError ? m.error.message : m.isError ? "Request failed." : null;

    return (
        <div className="flex flex-col gap-4">
            <StatusBadge status={draft.status} version={draft.version} approvedVersion={approvedVersion} />

            {draft.status === "invalid" || draft.status === "error" ? (
                <Banner tone="rose" icon={<AlertTriangle className="size-4" />}>
                    Research {draft.status === "error" ? "failed" : "returned an invalid result"}: {draft.error || "unknown"}. Fix the fields below and save.
                </Banner>
            ) : null}

            {flags.length > 0 ? (
                <Banner tone="amber" icon={<AlertTriangle className="size-4" />}>
                    <span className="font-semibold">Valenced language found.</span> A description must stay neutral.
                    Reword to remove: {flags.map((f) => `“${f.term}”`).join(", ")}. Save to re-check.
                </Banner>
            ) : null}

            <Field label="Game">
                <input className={inputClass} value={profile.game} onChange={(e) => edit({ game: e.target.value })} />
            </Field>
            <Field label="Developer">
                <input className={inputClass} value={profile.developer} onChange={(e) => edit({ developer: e.target.value })} />
            </Field>
            <Field label="Genre" hint="Factual genre/format.">
                <input className={inputClass} value={profile.genre} onChange={(e) => edit({ genre: e.target.value })} />
            </Field>
            <Field label="Core loop" hint="What the player does moment to moment.">
                <textarea rows={2} className={inputClass} value={profile.core_loop} onChange={(e) => edit({ core_loop: e.target.value })} />
            </Field>

            <fieldset className="flex flex-col gap-3 rounded-xl border border-slate-100 p-3">
                <legend className="px-1 text-xs font-semibold text-slate-500">Business model</legend>
                <Field label="Current state">
                    <input className={inputClass} value={profile.business_model.current_state} onChange={(e) => editBM({ current_state: e.target.value })} />
                </Field>
                <Field label="At full release" hint="Leave empty if the model is the same.">
                    <input className={inputClass} value={profile.business_model.at_full_release ?? ""} onChange={(e) => editBM({ at_full_release: e.target.value || null })} />
                </Field>
                <Field label="Real money buys" hint="State plainly whether gameplay power/progression is sold.">
                    <textarea rows={2} className={inputClass} value={profile.business_model.real_money_scope} onChange={(e) => editBM({ real_money_scope: e.target.value })} />
                </Field>
                <Field label="Anticipated / unconfirmed" hint="Reported-but-unconfirmed features, to verify. Leave empty if none.">
                    <textarea rows={2} className={inputClass} value={profile.business_model.anticipated_unconfirmed ?? ""} onChange={(e) => editBM({ anticipated_unconfirmed: e.target.value || null })} />
                </Field>
            </fieldset>

            <ResourceEditor
                resources={profile.resources}
                onChange={(resources) => edit({ resources })}
            />

            <StringListEditor
                label="Notable mechanics"
                hint="Each stated as existing, not judged."
                items={profile.notable_mechanics}
                onChange={(notable_mechanics) => edit({ notable_mechanics })}
            />

            <StringListEditor
                label="Sources"
                hint="URLs the description is grounded in."
                items={profile.sources}
                onChange={(sources) => edit({ sources })}
            />

            <RenderedPreview text={draft.renderedText} stale={dirty} />

            {errorMessage(update) ? <Banner tone="rose">{errorMessage(update)}</Banner> : null}
            {errorMessage(approve) ? <Banner tone="rose">{errorMessage(approve)}</Banner> : null}

            <div className="flex flex-wrap items-center gap-2 border-t border-slate-100 pt-4">
                <button
                    type="button"
                    onClick={save}
                    disabled={update.isPending || !dirty}
                    className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 px-3 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                >
                    <Save className="size-4" /> {update.isPending ? "Saving…" : "Save draft"}
                </button>
                <button
                    type="button"
                    onClick={() => approve.mutate(draft.id)}
                    disabled={!canApprove || approve.isPending}
                    title={dirty ? "Save your edits first" : flags.length > 0 ? "Clear the valenced language first" : "Freeze this as the approved version"}
                    className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                    <Check className="size-4" /> {approve.isPending ? "Approving…" : "Approve & freeze"}
                </button>
                <ResearchButton label="Re-research" pending={researchPending} done={researchDone} onClick={onResearch} subtle />
            </div>
        </div>
    );
}

function ResourceEditor({ resources, onChange }: { resources: Resource[]; onChange: (r: Resource[]) => void }) {
    const set = (i: number, patch: Partial<Resource>) =>
        onChange(resources.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));
    const add = () =>
        onChange([...resources, { name: "", acquisition: "earned-through-play", real_money_purchasable: false, purpose: "" }]);
    const remove = (i: number) => onChange(resources.filter((_, idx) => idx !== i));

    return (
        <fieldset className="flex flex-col gap-3 rounded-xl border border-slate-100 p-3">
            <legend className="px-1 text-xs font-semibold text-slate-500">Resources & currencies</legend>
            {resources.map((r, i) => (
                <div key={i} className="flex flex-col gap-2 rounded-lg border border-slate-100 bg-slate-50/50 p-2.5">
                    <div className="flex items-center gap-2">
                        <input className={`${inputClass} flex-1`} placeholder="name" value={r.name} onChange={(e) => set(i, { name: e.target.value })} />
                        <button type="button" onClick={() => remove(i)} aria-label="Remove" className="grid size-8 shrink-0 place-items-center rounded-lg text-slate-400 hover:bg-white hover:text-slate-700">
                            <X className="size-4" />
                        </button>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                        <select className={inputClass} value={r.acquisition} onChange={(e) => set(i, { acquisition: e.target.value as Acquisition })}>
                            <option value="earned-through-play">earned-through-play</option>
                            <option value="purchasable">purchasable</option>
                            <option value="both">both</option>
                        </select>
                        <label className="inline-flex items-center gap-1.5 text-xs text-slate-600">
                            <input type="checkbox" checked={r.real_money_purchasable} onChange={(e) => set(i, { real_money_purchasable: e.target.checked })} />
                            real-money purchasable
                        </label>
                    </div>
                    <input className={inputClass} placeholder="purpose — say if it's an economy vs premium/store currency" value={r.purpose} onChange={(e) => set(i, { purpose: e.target.value })} />
                </div>
            ))}
            <AddButton label="Add resource" onClick={add} />
        </fieldset>
    );
}

function StringListEditor({ label, hint, items, onChange }: { label: string; hint?: string; items: string[]; onChange: (v: string[]) => void }) {
    return (
        <Field label={label} hint={hint}>
            <div className="flex flex-col gap-2">
                {items.map((v, i) => (
                    <div key={i} className="flex items-center gap-2">
                        <input className={`${inputClass} flex-1`} value={v} onChange={(e) => onChange(items.map((x, idx) => (idx === i ? e.target.value : x)))} />
                        <button type="button" onClick={() => onChange(items.filter((_, idx) => idx !== i))} aria-label="Remove" className="grid size-8 shrink-0 place-items-center rounded-lg text-slate-400 hover:bg-slate-50 hover:text-slate-700">
                            <X className="size-4" />
                        </button>
                    </div>
                ))}
                <AddButton label={`Add ${label.toLowerCase()}`} onClick={() => onChange([...items, ""])} />
            </div>
        </Field>
    );
}

function AddButton({ label, onClick }: { label: string; onClick: () => void }) {
    return (
        <button type="button" onClick={onClick} className="inline-flex w-fit items-center gap-1.5 rounded-lg border border-slate-200 px-2.5 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50">
            <Plus className="size-3.5" /> {label}
        </button>
    );
}

function RenderedPreview({ text, stale }: { text: string; stale?: boolean }) {
    return (
        <div>
            <p className="mb-1 text-xs font-medium text-slate-500">
                Injected text {stale ? <span className="text-amber-600">(save to refresh)</span> : <span className="text-slate-400">(exactly what the panel sees)</span>}
            </p>
            <pre className="max-h-48 overflow-y-auto whitespace-pre-wrap rounded-lg border border-slate-100 bg-slate-50/60 px-3 py-2 text-xs text-slate-600">
                {text || "—"}
            </pre>
        </div>
    );
}

function ResearchButton({ label, pending, done, onClick, subtle }: { label: string; pending: boolean; done: boolean; onClick: () => void; subtle?: boolean }) {
    return (
        <div className="flex flex-col gap-1">
            <button
                type="button"
                onClick={onClick}
                disabled={pending}
                className={
                    subtle
                        ? "inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium text-slate-500 transition-colors hover:bg-slate-50 hover:text-slate-700 disabled:opacity-50"
                        : "inline-flex w-fit items-center gap-1.5 rounded-lg bg-slate-900 px-3.5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-slate-700 disabled:opacity-60"
                }
            >
                <RefreshCw className={`size-4 ${pending ? "animate-spin" : ""}`} /> {pending ? "Queuing…" : label}
            </button>
            {done ? <span className="text-xs text-slate-400">Research queued — it appears here once the worker finishes. Reopen to refresh.</span> : null}
        </div>
    );
}

function StatusBadge({ status, version, approvedVersion }: { status: string; version?: number; approvedVersion?: number }) {
    const tones: Record<string, string> = {
        draft: "border-amber-200 bg-amber-50 text-amber-700",
        approved: "border-emerald-200 bg-emerald-50 text-emerald-700",
        invalid: "border-rose-200 bg-rose-50 text-rose-700",
        error: "border-rose-200 bg-rose-50 text-rose-700",
    };
    return (
        <div className="flex flex-wrap items-center gap-2">
            <span className={`rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize ${tones[status] ?? "border-slate-200 bg-slate-50 text-slate-600"}`}>
                {status}{version ? ` v${version}` : ""}
            </span>
            {approvedVersion && status !== "approved" ? (
                <span className="text-xs text-slate-400">replaces approved v{approvedVersion} on approval</span>
            ) : null}
        </div>
    );
}

function Banner({ tone, icon, children }: { tone: "rose" | "amber"; icon?: React.ReactNode; children: React.ReactNode }) {
    const cls = tone === "rose" ? "border-rose-200 bg-rose-50 text-rose-700" : "border-amber-200 bg-amber-50 text-amber-800";
    return (
        <div className={`flex items-start gap-2 rounded-lg border px-3 py-2 text-xs ${cls}`}>
            {icon ? <span className="mt-0.5 shrink-0">{icon}</span> : null}
            <p>{children}</p>
        </div>
    );
}
