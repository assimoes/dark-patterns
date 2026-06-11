"use client";

import Link from "next/link";
import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { useListPopulations } from "@/hooks/useListPopulations";
import { fmt } from "@/lib/types";

// Populations table.

export default function BrowsePopulationsPage() {
    const { data, isPending, isError, refetch } = useListPopulations();
    const populations = data ?? [];

    return (
        <BrowsePanel
            title="Populations"
            subtitle="Materialised populations of reviews. Click a row to inspect it."
            badge={data ? <Badge>{populations.length} populations</Badge> : undefined}
        >
            {isPending ? (
                <TableSkeleton columns={5} />
            ) : isError ? (
                <TableError onRetry={() => refetch()} />
            ) : populations.length === 0 ? (
                <TableEmpty message="No populations yet." />
            ) : (
                <table className="w-full text-sm">
                    <TableHead
                        columns={["ID", "Label", "Modality", "Individuals", "Created"]}
                    />
                    <tbody>
                        {populations.map((p) => (
                            <tr
                                key={p.id}
                                className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60"
                            >
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                    <Link
                                        href={`/operations/browse/populations/${p.id}`}
                                        className="block"
                                    >
                                        #{p.id}
                                    </Link>
                                </td>
                                <td className="px-3 py-2.5">
                                    <Link
                                        href={`/operations/browse/populations/${p.id}`}
                                        className="block font-medium text-slate-700 hover:underline"
                                    >
                                        {p.label}
                                    </Link>
                                </td>
                                <td className="px-3 py-2.5 text-slate-500">{p.modality}</td>
                                <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">
                                    {fmt(p.individuals)}
                                </td>
                                <td className="px-3 py-2.5 text-slate-500">
                                    {new Date(p.createdAt).toLocaleDateString("en-US")}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </BrowsePanel>
    );
}