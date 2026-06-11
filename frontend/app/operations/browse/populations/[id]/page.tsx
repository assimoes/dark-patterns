"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { usePopulation } from "@/hooks/usePopulation";
import { fmt } from "@/lib/types";

// population detail.
export default function PopulationDetailPage() {
    const params = useParams<{ id: string }>();
    const parsed = Number(params.id);
    const id = Number.isFinite(parsed) ? parsed : null;

    const { data, isPending, isError, refetch } = usePopulation(id);

    return (
        <div className="flex flex-col gap-6">
            <Link
                href="/operations/browse/populations"
                className="text-sm font-medium text-slate-500 hover:text-slate-700"
            >
                ← All populations
            </Link>

            <BrowsePanel
                title={data ? data.label : `Population #${params.id}`}
                subtitle={
                    data
                        ? `${data.modality} · created ${new Date(
                            data.createdAt,
                        ).toLocaleDateString("en-US")}`
                        : "Loading…"
                }
                badge={data ? <Badge>#{data.id}</Badge> : undefined}
            >
                {isPending ? (
                    <TableSkeleton columns={3} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : (
                    <div className="flex flex-col gap-6">
                        <div>
                            <h3 className="mb-2 text-xs font-medium text-slate-400">
                                Per-game coverage
                            </h3>
                            {data.perGame.length === 0 ? (
                                <TableEmpty message="No games in this population." />
                            ) : (
                                <table className="w-full text-sm">
                                    <TableHead columns={["Game", "Reviews", "Annotated"]} />
                                    <tbody>
                                        {data.perGame.map((row) => (
                                            <tr
                                                key={row.gameId}
                                                className="border-b border-slate-50 last:border-0"
                                            >
                                                <td className="px-3 py-2.5 font-medium text-slate-700">
                                                    {row.name}
                                                </td>
                                                <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">
                                                    {fmt(row.reviews)}
                                                </td>
                                                <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">
                                                    {fmt(row.annotated)}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            )}
                        </div>

                        <div>
                            <h3 className="mb-2 text-xs font-medium text-slate-400">
                                Runs on this population
                            </h3>
                            {data.runs.length === 0 ? (
                                <TableEmpty message="No runs opened over this population yet." />
                            ) : (
                                <table className="w-full text-sm">
                                    <TableHead columns={["ID", "Label", "Type"]} />
                                    <tbody>
                                        {data.runs.map((run) => (
                                            <tr
                                                key={run.id}
                                                className="border-b border-slate-50 last:border-0 hover:bg-slate-50/60"
                                            >
                                                <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                                    <Link
                                                        href={`/operations/browse/runs/${run.id}`}
                                                        className="block"
                                                    >
                                                        #{run.id}
                                                    </Link>
                                                </td>
                                                <td className="px-3 py-2.5">
                                                    <Link
                                                        href={`/operations/browse/runs/${run.id}`}
                                                        className="block font-medium text-slate-700 hover:underline"
                                                    >
                                                        {run.label}
                                                    </Link>
                                                </td>
                                                <td className="px-3 py-2.5 text-slate-500">
                                                    {run.runType}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            )}
                        </div>
                    </div>
                )}
            </BrowsePanel>
        </div>
    );
}