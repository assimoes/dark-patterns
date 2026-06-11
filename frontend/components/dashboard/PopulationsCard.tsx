import { Users } from "lucide-react";
import { Card, Badge } from "@/components/ui/Card";
import { fmt } from "@/lib/types";
import { Game, PopulationStat } from "@/lib/types";

type PopulationsCardArgs = {
    games: Game[];
    populations: PopulationStat[];
}

export function PopulationsCard({
    games,
    populations
}: PopulationsCardArgs) {

    const gameById = (id: string): Game | undefined => games.find((g) => g.id === id);

    const total = populations.reduce((sum, p) => sum + p.individuals, 0)
    const max = Math.max(1, ...populations.map((p) => p.individuals));

    return (
        <Card
            icon={<Users className="size-5" />}
            iconClass="bg-indigo-50 text-indigo-600"
            title="Populations"
            subtitle="Curated individuals per game"
            badge={<Badge>{populations.length} populations</Badge>}
        >
            <div className="mb-5 flex items-end gap-2">
                <span className="text-4xl font-semibold tracking-tight text-slate-900 tabular-nums">
                    {fmt(total)}
                </span>
                <span className="pb-1.5 text-sm text-slate-500">individuals</span>
            </div>

            <ul className="space-y-3.5">
                {populations.map((p) => {
                    const g = gameById(p.gameId);
                    const color = g?.color ?? '#94a3b8'
                    const name = g?.name ?? p.gameId
                    return (
                        <li key={p.gameId}>
                            <div className="mb-1.5 flex items-center justify-between text-sm">
                                <span className="flex items-center gap-2 font-medium text-slate-700">
                                    <span
                                        className="size-2.5 shrink-0 rounded-full"
                                        style={{ backgroundColor: color }}
                                    />
                                    {name}
                                </span>
                                <span className="text-slate-900 tabular-nums">{fmt(p.individuals)}</span>
                            </div>
                            <div className="h-2 overflow-hidden rounded-full bg-slate-100">
                                <div
                                    className="h-full rounded-full transition-[width] duration-500"
                                    style={{ width: `${(p.individuals / max) * 100}%`, backgroundColor: color }}
                                />
                            </div>
                        </li>
                    );
                })}
            </ul>
        </Card>
    );
}