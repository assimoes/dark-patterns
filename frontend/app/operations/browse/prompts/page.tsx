"use client";

import { useState } from "react";
import { Pencil, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    NewButton,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { Drawer } from "@/components/operations/Drawer";
import { ConfirmDialog } from "@/components/operations/ConfirmDialog";
import { RowActions } from "@/components/operations/RowActions";
import { PromptForm } from "@/components/operations/forms/PromptForm";
import { useListPrompts } from "@/hooks/useListPrompts";
import { useDeletePrompt } from "@/hooks/useEntityMutations";
import { ApiError } from "@/lib/api/utils";
import type { PromptRow } from "@/lib/types";

type DrawerState = { mode: "new" } | { mode: "edit"; prompt: PromptRow };

export default function BrowsePromptsPage() {
    const { data, isPending, isError, refetch } = useListPrompts();
    const prompts = data ?? [];

    const [drawer, setDrawer] = useState<DrawerState | null>(null);
    const [toDelete, setToDelete] = useState<PromptRow | null>(null);
    const del = useDeletePrompt();

    const frozen = (p: PromptRow) => (p.runCount > 0 ? `Frozen: ${p.runCount} run(s) use this prompt.` : undefined);

    return (
        <>
            <BrowsePanel
                title="Prompts"
                subtitle="Prompt templates by version and modality."
                badge={
                    <div className="flex items-center gap-3">
                        {data ? <Badge>{prompts.length} prompts</Badge> : null}
                        <NewButton onClick={() => setDrawer({ mode: "new" })} />
                    </div>
                }
            >
                {isPending ? (
                    <TableSkeleton columns={5} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : prompts.length === 0 ? (
                    <TableEmpty message="No prompts available." />
                ) : (
                    <table className="w-full text-sm">
                        <TableHead columns={["ID", "Name", "Version", "Modality", ""]} />
                        <tbody>
                            {prompts.map((p) => (
                                <tr key={p.id} className="border-b border-slate-50 last:border-0">
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-400">#{p.id}</td>
                                    <td className="px-3 py-2.5 font-medium text-slate-700">{p.name}</td>
                                    <td className="px-3 py-2.5 tabular-nums text-slate-500">v{p.version}</td>
                                    <td className="px-3 py-2.5 text-slate-500">{p.modality}</td>
                                    <td className="px-3 py-2.5">
                                        <RowActions
                                            actions={[
                                                {
                                                    icon: <Pencil className="size-4" />,
                                                    label: "Edit",
                                                    onClick: () => setDrawer({ mode: "edit", prompt: p }),
                                                    disabledReason: frozen(p),
                                                },
                                                {
                                                    icon: <Trash2 className="size-4" />,
                                                    label: "Delete",
                                                    danger: true,
                                                    onClick: () => setToDelete(p),
                                                    disabledReason: frozen(p),
                                                },
                                            ]}
                                        />
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </BrowsePanel>

            <Drawer
                open={drawer !== null}
                onClose={() => setDrawer(null)}
                title={drawer?.mode === "edit" ? "Edit prompt" : "Add prompt"}
                subtitle={drawer?.mode === "edit" ? drawer.prompt.name : "Seed a new prompt version."}
            >
                {drawer?.mode === "new" ? <PromptForm onDone={() => setDrawer(null)} /> : null}
                {drawer?.mode === "edit" ? <PromptForm editingId={drawer.prompt.id} onDone={() => setDrawer(null)} /> : null}
            </Drawer>

            <ConfirmDialog
                open={toDelete !== null}
                title={`Delete ${toDelete?.name} v${toDelete?.version}?`}
                body="This removes the prompt. No run uses it."
                confirmLabel="Delete"
                pending={del.isPending}
                error={del.error instanceof ApiError ? del.error.message : null}
                onConfirm={() => toDelete && del.mutate(String(toDelete.id), { onSuccess: () => setToDelete(null) })}
                onCancel={() => { setToDelete(null); del.reset(); }}
            />
        </>
    );
}
