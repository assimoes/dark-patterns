"use client";

import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { useListAnnotators } from "@/hooks/useListAnnotators";

// Annotators table
export default function BrowseAnnotatorsPage() {
    const { data, isPending, isError, refetch } = useListAnnotators();
    const annotators = data ?? [];

    return (
        <BrowsePanel
            title="Annotators"
            subtitle="Human auditors and llm panel members."
            badge={data ? <Badge>{annotators.length} annotators</Badge> : undefined}
        >
            {isPending ? (
                <TableSkeleton columns={4} />
            ) : isError ? (
                <TableError onRetry={() => refetch()} />
            ) : annotators.length === 0 ? (
                <TableEmpty message="No annotators yet." />
            ) : (
                <table className="w-full text-sm">
                    <TableHead columns={["ID", "Kind", "Label", "Model"]} />
                    <tbody>
                        {annotators.map((a) => (
                            <tr
                                key={a.id}
                                className="border-b border-slate-50 last:border-0"
                            >
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                    #{a.id}
                                </td>
                                <td className="px-3 py-2.5">
                                    <span
                                        className={
                                            a.kind === "llm"
                                                ? "inline-flex items-center gap-1.5 text-violet-600"
                                                : "inline-flex items-center gap-1.5 text-amber-600"
                                        }
                                    >
                                        <span
                                            className={
                                                a.kind === "llm"
                                                    ? "size-1.5 rounded-full bg-violet-500"
                                                    : "size-1.5 rounded-full bg-amber-500"
                                            }
                                        />
                                        {a.kind}
                                    </span>
                                </td>
                                <td className="px-3 py-2.5 font-medium text-slate-700">
                                    {a.label}
                                </td>
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-500">
                                    {a.model ?? "—"}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </BrowsePanel>
    );
}