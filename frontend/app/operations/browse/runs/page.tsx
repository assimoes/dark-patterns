"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ListChecks, Shuffle, Trash2 } from "lucide-react";
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
import { RunForm } from "@/components/operations/forms/RunForm";
import { AdjSampleForm } from "@/components/operations/forms/AdjSampleForm";
import { useListRuns } from "@/hooks/useListRuns";
import { useDeleteRun } from "@/hooks/useEntityMutations";
import { useEnqueueAnnotations } from "@/hooks/useEnqueueAnnotations";
import { api } from "@/lib/api";
import { ApiError } from "@/lib/api/utils";
import type { RunSummary } from "@/lib/types";

type DrawerState = { mode: "new" } | { mode: "adj"; run: RunSummary };

export default function BrowseRunsPage() {
    const { data, isPending, isError, refetch } = useListRuns();
    const runs = data ?? [];

    const [drawer, setDrawer] = useState<DrawerState | null>(null);
    const [toDelete, setToDelete] = useState<RunSummary | null>(null);
    const [toEnqueue, setToEnqueue] = useState<RunSummary | null>(null);
    const del = useDeleteRun();
    const enqueue = useEnqueueAnnotations();

    const impact = useQuery({
        queryKey: ["runImpact", toDelete?.id],
        queryFn: () => api.operations.runImpact(String(toDelete!.id)),
        enabled: toDelete !== null,
    });

    return (
        <>
            <BrowsePanel
                title="Runs"
                subtitle="Annotation runs. Click a row to inspect its panel."
                badge={
                    <div className="flex items-center gap-3">
                        {data ? <Badge>{runs.length} runs</Badge> : null}
                        <NewButton onClick={() => setDrawer({ mode: "new" })} />
                    </div>
                }
            >
                {isPending ? (
                    <TableSkeleton columns={8} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : runs.length === 0 ? (
                    <TableEmpty message="No runs yet." />
                ) : (
                    <table className="w-full text-sm">
                        <TableHead columns={["ID", "Type", "Population", "Prompt", "Taxonomy", "Panel", "Created", ""]} />
                        <tbody>
                            {runs.map((r) => (
                                <tr key={r.id} className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60">
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                        <Link href={`/operations/browse/runs/${r.id}`} className="block">#{r.id}</Link>
                                    </td>
                                    <td className="px-3 py-2.5">
                                        <Link href={`/operations/browse/runs/${r.id}`} className="block font-medium text-slate-700 hover:underline">
                                            {r.runType}
                                        </Link>
                                    </td>
                                    <td className="px-3 py-2.5 text-slate-500">{r.population}</td>
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-500">#{r.promptId}</td>
                                    <td className="px-3 py-2.5 tabular-nums text-slate-500">{r.taxonomyVersion ?? "—"}</td>
                                    <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">{r.panelSize}</td>
                                    <td className="px-3 py-2.5 text-slate-500">{new Date(r.createdAt).toLocaleDateString("en-US")}</td>
                                    <td className="px-3 py-2.5">
                                        <RowActions
                                            actions={[
                                                {
                                                    icon: <ListChecks className="size-4" />,
                                                    label: "Enqueue annotations",
                                                    onClick: () => setToEnqueue(r),
                                                    disabledReason: r.runType !== "llm_panel" ? "Only llm_panel runs annotate." : undefined,
                                                },
                                                {
                                                    icon: <Shuffle className="size-4" />,
                                                    label: "Create adjudication sample",
                                                    onClick: () => setDrawer({ mode: "adj", run: r }),
                                                    disabledReason: r.runType !== "llm_panel" ? "Only llm_panel runs can be sampled." : undefined,
                                                },
                                                { icon: <Trash2 className="size-4" />, label: "Delete", danger: true, onClick: () => setToDelete(r) },
                                            ]}
                                        />
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </BrowsePanel>

            <Drawer open={drawer?.mode === "new"} onClose={() => setDrawer(null)} title="Create run" subtitle="Open an annotation run over a population.">
                {drawer?.mode === "new" ? <RunForm onDone={() => setDrawer(null)} /> : null}
            </Drawer>

            <Drawer
                open={drawer?.mode === "adj"}
                onClose={() => setDrawer(null)}
                title="Create adjudication sample"
                subtitle={drawer?.mode === "adj" ? `run #${drawer.run.id}` : undefined}
            >
                {drawer?.mode === "adj" ? <AdjSampleForm presetPanelRunId={String(drawer.run.id)} /> : null}
            </Drawer>

            <ConfirmDialog
                open={toEnqueue !== null}
                title={`Enqueue annotations for run #${toEnqueue?.id}?`}
                body="One annotation job per (review, panel member) is queued. Existing annotations are not re-run."
                confirmLabel="Enqueue"
                pendingLabel="Enqueuing…"
                pending={enqueue.isPending}
                error={enqueue.error instanceof ApiError ? enqueue.error.message : null}
                onConfirm={() => toEnqueue && enqueue.mutate(String(toEnqueue.id), { onSuccess: () => setToEnqueue(null) })}
                onCancel={() => { setToEnqueue(null); enqueue.reset(); }}
            />

            <ConfirmDialog
                open={toDelete !== null}
                title={`Delete run #${toDelete?.id}?`}
                body={impact.isPending ? "Computing what this removes…" : "This cascades to everything below. The population and other runs stay."}
                items={
                    impact.data
                        ? [
                            { label: "Annotations", value: impact.data.annotations },
                            { label: "Adjudication samples", value: impact.data.samples },
                            { label: "Adjudications", value: impact.data.adjudications },
                        ]
                        : undefined
                }
                confirmLabel="Delete all"
                pending={del.isPending}
                error={del.error instanceof ApiError ? del.error.message : null}
                onConfirm={() => toDelete && del.mutate(String(toDelete.id), { onSuccess: () => setToDelete(null) })}
                onCancel={() => { setToDelete(null); del.reset(); }}
            />
        </>
    );
}
