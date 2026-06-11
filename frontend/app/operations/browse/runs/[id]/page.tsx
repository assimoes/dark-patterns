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
import { useRun } from "@/hooks/useRun";

// Run detail
export default function RunDetailPage() {
    const params = useParams<{ id: string }>();
    const parsed = Number(params.id);
    const id = Number.isFinite(parsed) ? parsed : null;

    const { data, isPending, isError, refetch } = useRun(id);

    return (
        <div className="flex flex-col gap-6">
            <Link
                href="/operations/browse/runs"
                className="text-sm font-medium text-slate-500 hover:text-slate-700"
            >
                ← All runs
            </Link>

            <BrowsePanel
                title={data ? `Run #${data.id}` : `Run #${params.id}`}
                subtitle={
                    data
                        ? `${data.runType} · ${data.population} · created ${new Date(
                            data.createdAt,
                        ).toLocaleDateString("en-US")}`
                        : "Loading…"
                }
                badge={
                    data ? (
                        <Badge
                            className={
                                data.hasSample
                                    ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                                    : ""
                            }
                        >
                            {data.hasSample ? "sample drawn" : "no sample"}
                        </Badge>
                    ) : undefined
                }
            >
                {isPending ? (
                    <TableSkeleton columns={2} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : (
                    <div className="flex flex-col gap-6">
                        <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm sm:grid-cols-3">
                            <Meta label="Run type" value={data.runType} />
                            <Meta
                                label="Population"
                                value={`${data.population} (#${data.populationId})`}
                            />
                            <Meta label="Prompt" value={`#${data.promptId}`} />
                            <Meta
                                label="Taxonomy"
                                value={
                                    data.taxonomyVersion === null
                                        ? "—"
                                        : String(data.taxonomyVersion)
                                }
                            />
                            <Meta
                                label="Sample"
                                value={data.hasSample ? "drawn" : "not drawn"}
                            />
                        </dl>

                        <div>
                            <h3 className="mb-2 text-xs font-medium text-slate-400">
                                Panel members
                            </h3>
                            {data.panel.length === 0 ? (
                                <TableEmpty message="No panel members on this run." />
                            ) : (
                                <table className="w-full text-sm">
                                    <TableHead columns={["Kind", "Label"]} />
                                    <tbody>
                                        {data.panel.map((m, i) => (
                                            <tr
                                                key={`${m.kind}-${m.label}-${i}`}
                                                className="border-b border-slate-50 last:border-0"
                                            >
                                                <td className="px-3 py-2.5">
                                                    <span
                                                        className={
                                                            m.kind === "llm"
                                                                ? "inline-flex items-center gap-1.5 text-violet-600"
                                                                : "inline-flex items-center gap-1.5 text-amber-600"
                                                        }
                                                    >
                                                        <span
                                                            className={
                                                                m.kind === "llm"
                                                                    ? "size-1.5 rounded-full bg-violet-500"
                                                                    : "size-1.5 rounded-full bg-amber-500"
                                                            }
                                                        />
                                                        {m.kind}
                                                    </span>
                                                </td>
                                                <td className="px-3 py-2.5 font-medium text-slate-700">
                                                    {m.label}
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

// One label/value pair in the run's metadata grid.
function Meta({ label, value }: { label: string; value: string }) {
    return (
        <div>
            <dt className="text-xs font-medium text-slate-400">{label}</dt>
            <dd className="mt-0.5 font-medium text-slate-700">{value}</dd>
        </div>
    );
}