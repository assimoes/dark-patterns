"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

// The operator area's tab strip. one tab per entity; each page lists and manages (create/edit/delete).
const tabs = [
    { href: "/operations/browse/games", label: "Games" },
    { href: "/operations/browse/populations", label: "Populations" },
    { href: "/operations/browse/runs", label: "Runs" },
    { href: "/operations/browse/prompts", label: "Prompts" },
    { href: "/operations/browse/annotators", label: "Annotators" },
    { href: "/operations/browse/comparisons", label: "Comparisons" },
];

export function OperationsNav() {
    const pathname = usePathname();

    return (
        <nav className="flex flex-wrap gap-1.5">
            {tabs.map((tab) => {
                const active = pathname.startsWith(tab.href);
                return (
                    <Link
                        key={tab.href}
                        href={tab.href}
                        className={
                            active
                                ? "rounded-lg bg-slate-900 px-3 py-1.5 text-sm font-medium text-white transition-colors"
                                : "rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                        }
                    >
                        {tab.label}
                    </Link>
                );
            })}
        </nav>
    );
}