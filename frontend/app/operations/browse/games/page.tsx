"use client";

import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { useListGames } from "@/hooks/useListGames";
import { fmt } from "@/lib/types";

// Curated games table

export default function BrowseGamesPage() {
    const { data, isPending, isError, refetch } = useListGames();
    const games = data ?? [];

    return (
        <BrowsePanel
            title="Games"
            subtitle="Curated Steam games and their corpus coverage."
            badge={data ? <Badge>{games.length} games</Badge> : undefined}
        >
            {isPending ? (
                <TableSkeleton columns={5} />
            ) : isError ? (
                <TableError onRetry={() => refetch()} />
            ) : games.length === 0 ? (
                <TableEmpty message="No games registered yet." />
            ) : (
                <table className="w-full text-sm">
                    <TableHead
                        columns={["Game", "Short", "Monetization", "Reviews", "Annotated"]}
                    />
                    <tbody>
                        {games.map((g) => (
                            <tr
                                key={g.id}
                                className="border-b border-slate-50 last:border-0"
                            >
                                <td className="px-3 py-2.5">
                                    <span className="flex items-center gap-2">
                                        <span
                                            className="size-3 shrink-0 rounded-full"
                                            style={{ backgroundColor: g.color }}
                                        />
                                        <span className="font-medium text-slate-700">{g.name}</span>
                                    </span>
                                </td>
                                <td className="px-3 py-2.5 font-mono text-xs text-slate-500">
                                    {g.short}
                                </td>
                                <td className="px-3 py-2.5 text-slate-500">{g.monetization}</td>
                                <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">
                                    {fmt(g.reviews)}
                                </td>
                                <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">
                                    {fmt(g.annotated)}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </BrowsePanel>
    );
}