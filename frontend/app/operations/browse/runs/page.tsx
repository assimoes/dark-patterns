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
import { useListRuns } from "@/hooks/useListRuns";

// Runs table

export default function BrowseRunsPage() {
    const { data, isPending, isError, refetch } = useListRuns();
    const runs = data ?? [];

    return (
        <BrowsePanel
            title="Runs"
            subtitle="Annotation runs. Click a row to inspect its panel."
            badge={data ? <Badge>{runs.length} runs</Badge> : undefined}
        >
            {isPending ? (
                <TableSkeleton columns={7} />
            ) : isError ? (
                <TableError onRetry={() => refetch()} />
            ) : runs.length === 0 ? (
                <TableEmpty message="No runs yet." />
            ) : (
                <table className="w-full text-sm">
                    <TableHead
                        columns={[
                            "ID",
                            "Type",
                            "Population",
                            "Prompt",
                            "Taxonomy",
                            "Panel",
                            "Created",
                        ]}
                    />
                    <tbody>
                        {runs.map((r) => (
                            <tr
                                key={r.id}
                                className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60"
                            >
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                    <Link
                                        href={`/operations/browse/runs/${r.id}`}
                                        className="block"
                                    >
                                        #{r.id}
                                    </Link>
                                </td>
                                <td className="px-3 py-2.5">
                                    <Link
                                        href={`/operations/browse/runs/${r.id}`}
                                        className="block font-medium text-slate-700 hover:underline"
                                    >
                                        {r.runType}
                                    </Link>
                                </td>
                                <td className="px-3 py-2.5 text-slate-500">{r.population}</td>
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-500">
                                    #{r.promptId}
                                </td>
                                <td className="px-3 py-2.5 tabular-nums text-slate-500">
                                    {r.taxonomyVersion ?? "—"}
                                </td>
                                <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">
                                    {r.panelSize}
                                </td>
                                <td className="px-3 py-2.5 text-slate-500">
                                    {new Date(r.createdAt).toLocaleDateString("en-US")}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </BrowsePanel>
    );
}