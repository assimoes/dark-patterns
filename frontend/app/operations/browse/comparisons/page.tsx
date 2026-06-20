"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Trash2 } from "lucide-react";
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
import { ComparisonBuilder } from "@/components/operations/ComparisonBuilder";
import { useComparisons, useDeleteComparison } from "@/hooks/useComparisons";
import type { Comparison } from "@/lib/types";

export default function ComparisonsPage() {
    const { data, isPending, isError, refetch } = useComparisons();
    const comparisons = data ?? [];
    const router = useRouter();

    const [building, setBuilding] = useState(false);
    const [toDelete, setToDelete] = useState<Comparison | null>(null);
    const del = useDeleteComparison();

    return (
        <>
            <BrowsePanel
                title="Comparisons"
                subtitle="Compare two runs over the same population, model by model."
                badge={
                    <div className="flex items-center gap-3">
                        {data ? <Badge>{comparisons.length} comparisons</Badge> : null}
                        <NewButton onClick={() => setBuilding(true)} />
                    </div>
                }
            >
                {isPending ? (
                    <TableSkeleton columns={4} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : comparisons.length === 0 ? (
                    <TableEmpty message="No comparisons yet." />
                ) : (
                    <table className="w-full text-sm">
                        <TableHead columns={["Label", "Runs", "Created", ""]} />
                        <tbody>
                            {comparisons.map((c) => (
                                <tr key={c.id} className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60">
                                    <td className="px-3 py-2.5">
                                        <Link href={`/operations/browse/comparisons/${c.id}`} className="font-medium text-slate-700 hover:underline">
                                            {c.label}
                                        </Link>
                                    </td>
                                    <td className="px-3 py-2.5">
                                        <span className="inline-flex items-center gap-1.5 font-mono text-xs text-slate-500">
                                            #{c.runAId} <ArrowRight className="size-3" /> #{c.runBId}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2.5 text-slate-500">{new Date(c.createdAt).toLocaleDateString("en-US")}</td>
                                    <td className="px-3 py-2.5">
                                        <RowActions
                                            actions={[
                                                { icon: <Trash2 className="size-4" />, label: "Delete", danger: true, onClick: () => setToDelete(c) },
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
                open={building}
                onClose={() => setBuilding(false)}
                title="New comparison"
                subtitle="Pick two runs to compare on the reviews they both annotated."
            >
                {building ? (
                    <ComparisonBuilder
                        onDone={(id) => {
                            setBuilding(false);
                            router.push(`/operations/browse/comparisons/${id}`);
                        }}
                    />
                ) : null}
            </Drawer>

            <ConfirmDialog
                open={toDelete !== null}
                title={`Delete ${toDelete?.label}?`}
                body="This removes the saved comparison. The runs and their annotations are untouched."
                confirmLabel="Delete"
                pending={del.isPending}
                onConfirm={() => toDelete && del.mutate(toDelete.id, { onSuccess: () => setToDelete(null) })}
                onCancel={() => setToDelete(null)}
            />
        </>
    );
}
