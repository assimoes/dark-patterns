import Link from "next/link";
import { Gamepad2, Layers, Users, Play, FileText } from "lucide-react";

// Browse index
const sections = [
    {
        href: "/operations/browse/games",
        title: "Games",
        description: "Curated Steam games with corpus coverage.",
        icon: <Gamepad2 className="size-5" />,
        iconClass: "bg-amber-50 text-amber-600",
    },
    {
        href: "/operations/browse/populations",
        title: "Populations",
        description: "Materialised populations and their per-game coverage.",
        icon: <Layers className="size-5" />,
        iconClass: "bg-violet-50 text-violet-600",
    },
    {
        href: "/operations/browse/annotators",
        title: "Annotators",
        description: "Human auditors and llm panel members.",
        icon: <Users className="size-5" />,
        iconClass: "bg-sky-50 text-sky-600",
    },
    {
        href: "/operations/browse/runs",
        title: "Runs",
        description: "Annotation runs with their panel and sample state.",
        icon: <Play className="size-5" />,
        iconClass: "bg-emerald-50 text-emerald-600",
    },
    {
        href: "/operations/browse/prompts",
        title: "Prompts",
        description: "Prompt templates by version and modality.",
        icon: <FileText className="size-5" />,
        iconClass: "bg-slate-100 text-slate-700",
    },
];

export default function BrowseIndexPage() {
    return (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {sections.map((section) => (
                <Link
                    key={section.href}
                    href={section.href}
                    className="flex flex-col gap-3 rounded-2xl border border-slate-200/80 bg-white/90 px-5 py-5 shadow-sm ring-1 ring-slate-900/[0.02] transition-colors hover:border-slate-300 hover:bg-white"
                >
                    <span
                        className={`grid size-10 place-items-center rounded-xl ${section.iconClass}`}
                    >
                        {section.icon}
                    </span>
                    <div>
                        <h2 className="text-sm font-semibold tracking-tight text-slate-900">
                            {section.title}
                        </h2>
                        <p className="mt-1 text-xs text-slate-500">{section.description}</p>
                    </div>
                </Link>
            ))}
        </div>
    );
}