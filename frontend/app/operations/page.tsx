import Link from "next/link";
import {
    Gamepad2,
    Users,
    Layers,
    Play,
    Download,
    ListChecks,
    Shuffle,
    Table2,
} from "lucide-react";

// The operator index
const actions = [
    {
        href: "/operations/browse",
        title: "Browse",
        description:
            "Inspect existing games, populations, annotators, runs, and prompts.",
        icon: <Table2 className="size-5" />,
        iconClass: "bg-slate-100 text-slate-700",
    },
    {
        href: "/operations/games",
        title: "Add game",
        description:
            "Register a curated Steam game with its presentation metadata.",
        icon: <Gamepad2 className="size-5" />,
        iconClass: "bg-amber-50 text-amber-600",
    },
    {
        href: "/operations/annotators",
        title: "Add annotator",
        description: "Add a human auditor or an llm panel member.",
        icon: <Users className="size-5" />,
        iconClass: "bg-sky-50 text-sky-600",
    },
    {
        href: "/operations/populations",
        title: "Create population",
        description: "Materialise a population of reviews from filter criteria.",
        icon: <Layers className="size-5" />,
        iconClass: "bg-violet-50 text-violet-600",
    },
    {
        href: "/operations/runs",
        title: "Create run",
        description: "Open an annotation run over a population with a prompt.",
        icon: <Play className="size-5" />,
        iconClass: "bg-emerald-50 text-emerald-600",
    },
    {
        href: "/operations/scrapes",
        title: "Enqueue scrape",
        description: "Queue a Steam reviews scrape for an app id.",
        icon: <Download className="size-5" />,
        iconClass: "bg-rose-50 text-rose-600",
    },
    {
        href: "/operations/annotations",
        title: "Enqueue annotations",
        description: "Kick off annotation jobs for an existing run.",
        icon: <ListChecks className="size-5" />,
        iconClass: "bg-slate-100 text-slate-700",
    },
    {
        href: "/operations/adjudication-sample",
        title: "Create adjudication sample",
        description: "Draw a stratified sample from a panel run to adjudicate.",
        icon: <Shuffle className="size-5" />,
        iconClass: "bg-violet-50 text-violet-600",
    },
];

export default function OperationsIndexPage() {
    return (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {actions.map((action) => (
                <Link
                    key={action.href}
                    href={action.href}
                    className="flex flex-col gap-3 rounded-2xl border border-slate-200/80 bg-white/90 px-5 py-5 shadow-sm ring-1 ring-slate-900/[0.02] transition-colors hover:border-slate-300 hover:bg-white"
                >
                    <span
                        className={`grid size-10 place-items-center rounded-xl ${action.iconClass}`}
                    >
                        {action.icon}
                    </span>
                    <div>
                        <h2 className="text-sm font-semibold tracking-tight text-slate-900">
                            {action.title}
                        </h2>
                        <p className="mt-1 text-xs text-slate-500">{action.description}</p>
                    </div>
                </Link>
            ))}
        </div>
    );
}