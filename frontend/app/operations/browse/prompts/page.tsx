"use client";

import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { useListPrompts } from "@/hooks/useListPrompts";

// Prompts table

export default function BrowsePromptsPage() {
    const { data, isPending, isError, refetch } = useListPrompts();
    const prompts = data ?? [];

    return (
        <BrowsePanel
            title="Prompts"
            subtitle="Prompt templates by version and modality."
            badge={data ? <Badge>{prompts.length} prompts</Badge> : undefined}
        >
            {isPending ? (
                <TableSkeleton columns={4} />
            ) : isError ? (
                <TableError onRetry={() => refetch()} />
            ) : prompts.length === 0 ? (
                <TableEmpty message="No prompts available." />
            ) : (
                <table className="w-full text-sm">
                    <TableHead columns={["ID", "Name", "Version", "Modality"]} />
                    <tbody>
                        {prompts.map((p) => (
                            <tr
                                key={p.id}
                                className="border-b border-slate-50 last:border-0"
                            >
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-400">
                                    #{p.id}
                                </td>
                                <td className="px-3 py-2.5 font-medium text-slate-700">
                                    {p.name}
                                </td>
                                <td className="px-3 py-2.5 tabular-nums text-slate-500">
                                    v{p.version}
                                </td>
                                <td className="px-3 py-2.5 text-slate-500">{p.modality}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </BrowsePanel>
    );
}