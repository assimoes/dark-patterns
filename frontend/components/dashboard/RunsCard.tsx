import { Layers, Cpu, User } from "lucide-react";
import { Card, Badge } from "@/components/ui/Card";
import { Run } from "@/lib/types";

type RunsCardArgs = {
    runs: Run[]
}

export function RunsCard({ runs }: RunsCardArgs) {

    const seats = runs.flatMap((r) => r.members)
    const llm = seats.filter((m) => m.kind === 'llm').length
    const human = seats.filter((m) => m.kind === 'human').length
    const total = seats.length || 1;

    return (
        <Card
            icon={<Layers className="size-5" />}
            iconClass="bg-violet-50 text-violet-600"
            title="Runs"
            subtitle="Panel members by type"
            badge={<Badge>{runs.length} runs</Badge>}
        >
            <div className="grid grid-cols-2 gap-3">
                <StatTile
                    icon={<Cpu className="size-4" />}
                    label="LLM members"
                    value={llm}
                    tone="text-violet-600 bg-violet-50"
                />
                <StatTile
                    icon={<User className="size-4" />}
                    label="Human members"
                    value={human}
                    tone="text-amber-600 bg-amber-50"
                />
            </div>

            {/* proportion bar: llm vs human across all runs */}
            <div className="mt-4">
                <div className="flex h-2 overflow-hidden rounded-full bg-slate-100">
                    <div className="bg-violet-500" style={{ width: `${(llm / total) * 100}%` }} />
                    <div className="bg-amber-500" style={{ width: `${(human / total) * 100}%` }} />
                </div>
                <p className="mt-2 text-xs text-slate-500">
                    <span className="font-medium text-slate-700">{total}</span> panel seats across{" "}
                    {runs.length} runs
                </p>
            </div>

            <ul className="mt-5 space-y-2">
                {runs.map((r) => {
                    const llmCount = r.members.filter((m) => m.kind === "llm").length;
                    const humanCount = r.members.filter((m) => m.kind === "human").length;
                    return (
                        <li
                            key={r.id}
                            className="flex items-center gap-3 rounded-xl border border-slate-100 bg-slate-50/50 px-3 py-2.5 text-sm"
                        >
                            <span className="font-mono text-xs text-slate-400">#{r.id}</span>
                            <span className="truncate font-medium text-slate-700">{r.label}</span>
                            <span className="hidden text-xs text-slate-400 sm:inline">{r.population}</span>
                            <span className="ml-auto flex items-center gap-3 text-xs text-slate-500">
                                {llmCount > 0 ? (
                                    <span className="flex items-center gap-1">
                                        <span className="size-1.5 rounded-full bg-violet-500" />
                                        {llmCount} llm
                                    </span>
                                ) : null}
                                {humanCount > 0 ? (
                                    <span className="flex items-center gap-1">
                                        <span className="size-1.5 rounded-full bg-amber-500" />
                                        {humanCount} human
                                    </span>
                                ) : null}
                            </span>
                        </li>
                    );
                })}
            </ul>
        </Card>
    );
}

type StatTitleArgs = {
    icon: React.ReactNode;
    label: string;
    value: number;
    tone: string;
}

function StatTile({
    icon,
    label,
    value,
    tone,
}: StatTitleArgs) {
    return (
        <div className="rounded-xl border border-slate-100 bg-white px-4 py-3.5">
            <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-slate-500">{label}</span>
                <span className={`grid size-6 place-items-center rounded-md ${tone}`}>{icon}</span>
            </div>
            <p className="mt-1.5 text-3xl font-semibold tracking-tight text-slate-900 tabular-nums">
                {value}
            </p>
        </div>
    );
}
