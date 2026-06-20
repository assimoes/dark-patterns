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
import { AnnotatorForm } from "@/components/operations/forms/AnnotatorForm";
import { useListAnnotators } from "@/hooks/useListAnnotators";
import { useDeleteAnnotator } from "@/hooks/useEntityMutations";
import { ApiError } from "@/lib/api/utils";
import type { AnnotatorRow } from "@/lib/types";

type DrawerState = { mode: "new" } | { mode: "edit"; annotator: AnnotatorRow };

export default function BrowseAnnotatorsPage() {
    const { data, isPending, isError, refetch } = useListAnnotators();
    const annotators = data ?? [];

    const [drawer, setDrawer] = useState<DrawerState | null>(null);
    const [toDelete, setToDelete] = useState<AnnotatorRow | null>(null);
    const del = useDeleteAnnotator();

    const frozen = (a: AnnotatorRow) => (a.refCount > 0 ? `Frozen: referenced by ${a.refCount} run(s)/annotation(s).` : undefined);

    return (
        <>
            <BrowsePanel
                title="Annotators"
                subtitle="Human auditors and llm panel members."
                badge={
                    <div className="flex items-center gap-3">
                        {data ? <Badge>{annotators.length} annotators</Badge> : null}
                        <NewButton onClick={() => setDrawer({ mode: "new" })} />
                    </div>
                }
            >
                {isPending ? (
                    <TableSkeleton columns={5} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : annotators.length === 0 ? (
                    <TableEmpty message="No annotators yet." />
                ) : (
                    <table className="w-full text-sm">
                        <TableHead columns={["ID", "Kind", "Label", "Model", ""]} />
                        <tbody>
                            {annotators.map((a) => (
                                <tr key={a.id} className="border-b border-slate-50 last:border-0">
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-400">#{a.id}</td>
                                    <td className="px-3 py-2.5">
                                        <span
                                            className={
                                                a.kind === "llm"
                                                    ? "inline-flex items-center gap-1.5 text-violet-600"
                                                    : "inline-flex items-center gap-1.5 text-amber-600"
                                            }
                                        >
                                            <span className={a.kind === "llm" ? "size-1.5 rounded-full bg-violet-500" : "size-1.5 rounded-full bg-amber-500"} />
                                            {a.kind}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2.5 font-medium text-slate-700">{a.label}</td>
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-500">{a.model ?? "—"}</td>
                                    <td className="px-3 py-2.5">
                                        <RowActions
                                            actions={[
                                                { icon: <Pencil className="size-4" />, label: "Edit label", onClick: () => setDrawer({ mode: "edit", annotator: a }) },
                                                {
                                                    icon: <Trash2 className="size-4" />,
                                                    label: "Delete",
                                                    danger: true,
                                                    onClick: () => setToDelete(a),
                                                    disabledReason: frozen(a),
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
                title={drawer?.mode === "edit" ? "Edit annotator" : "Add annotator"}
                subtitle={drawer?.mode === "edit" ? drawer.annotator.label : "Add a human auditor or llm panel member."}
            >
                {drawer?.mode === "new" ? <AnnotatorForm onDone={() => setDrawer(null)} /> : null}
                {drawer?.mode === "edit" ? <AnnotatorForm editing={drawer.annotator} onDone={() => setDrawer(null)} /> : null}
            </Drawer>

            <ConfirmDialog
                open={toDelete !== null}
                title={`Delete ${toDelete?.label}?`}
                body="This removes the annotator. Nothing references it."
                confirmLabel="Delete"
                pending={del.isPending}
                error={del.error instanceof ApiError ? del.error.message : null}
                onConfirm={() => toDelete && del.mutate(String(toDelete.id), { onSuccess: () => setToDelete(null) })}
                onCancel={() => { setToDelete(null); del.reset(); }}
            />
        </>
    );
}
