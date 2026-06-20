"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { Play, Trash2 } from "lucide-react";
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
import { PopulationForm } from "@/components/operations/forms/PopulationForm";
import { RunForm } from "@/components/operations/forms/RunForm";
import { useListPopulations } from "@/hooks/useListPopulations";
import { useDeletePopulation } from "@/hooks/useEntityMutations";
import { api } from "@/lib/api";
import { ApiError } from "@/lib/api/utils";
import { fmt } from "@/lib/types";
import type { PopulationSummary } from "@/lib/types";

type DrawerState = { mode: "new" } | { mode: "run"; population: PopulationSummary };

export default function BrowsePopulationsPage() {
    const { data, isPending, isError, refetch } = useListPopulations();
    const populations = data ?? [];

    const [drawer, setDrawer] = useState<DrawerState | null>(null);
    const [toDelete, setToDelete] = useState<PopulationSummary | null>(null);
    const del = useDeletePopulation();

    // the blast radius is fetched only when a delete is pending, then shown in the confirm dialog.
    const impact = useQuery({
        queryKey: ["populationImpact", toDelete?.id],
        queryFn: () => api.operations.populationImpact(String(toDelete!.id)),
        enabled: toDelete !== null,
    });

    return (
        <>
            <BrowsePanel
                title="Populations"
                subtitle="Frozen populations of reviews. Click a row to inspect it."
                badge={
                    <div className="flex items-center gap-3">
                        {data ? <Badge>{populations.length} populations</Badge> : null}
                        <NewButton onClick={() => setDrawer({ mode: "new" })} />
                    </div>
                }
            >
                {isPending ? (
                    <TableSkeleton columns={6} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : populations.length === 0 ? (
                    <TableEmpty message="No populations yet." />
                ) : (
                    <table className="w-full text-sm">
                        <TableHead columns={["ID", "Label", "Modality", "Individuals", "Created", ""]} />
                        <tbody>
                            {populations.map((p) => (
                                <tr key={p.id} className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60">
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                        <Link href={`/operations/browse/populations/${p.id}`} className="block">#{p.id}</Link>
                                    </td>
                                    <td className="px-3 py-2.5">
                                        <Link href={`/operations/browse/populations/${p.id}`} className="block font-medium text-slate-700 hover:underline">
                                            {p.label}
                                        </Link>
                                    </td>
                                    <td className="px-3 py-2.5 text-slate-500">{p.modality}</td>
                                    <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">{fmt(p.individuals)}</td>
                                    <td className="px-3 py-2.5 text-slate-500">{new Date(p.createdAt).toLocaleDateString("en-US")}</td>
                                    <td className="px-3 py-2.5">
                                        <RowActions
                                            actions={[
                                                { icon: <Play className="size-4" />, label: "Create run", onClick: () => setDrawer({ mode: "run", population: p }) },
                                                { icon: <Trash2 className="size-4" />, label: "Delete", danger: true, onClick: () => setToDelete(p) },
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
                open={drawer?.mode === "new"}
                onClose={() => setDrawer(null)}
                title="Create population"
                subtitle="Freeze a population from filter criteria."
            >
                {drawer?.mode === "new" ? <PopulationForm onDone={() => setDrawer(null)} /> : null}
            </Drawer>

            <Drawer
                open={drawer?.mode === "run"}
                onClose={() => setDrawer(null)}
                title="Create run"
                subtitle={drawer?.mode === "run" ? drawer.population.label : undefined}
            >
                {drawer?.mode === "run" ? (
                    <RunForm presetPopulationId={String(drawer.population.id)} onDone={() => setDrawer(null)} />
                ) : null}
            </Drawer>

            <ConfirmDialog
                open={toDelete !== null}
                title={`Delete ${toDelete?.label}?`}
                body={impact.isPending ? "Computing what this removes…" : "This cascades to everything below. It cannot be undone."}
                items={
                    impact.data
                        ? [
                            { label: "Individuals", value: impact.data.individuals },
                            { label: "Runs", value: impact.data.runs },
                            { label: "Annotations", value: impact.data.annotations },
                            { label: "Adjudication samples", value: impact.data.samples },
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
