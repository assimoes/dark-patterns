"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

// The operator area's tab strip.
const tabs = [
    { href: "/operations/browse", label: "Browse" },
    { href: "/operations/games", label: "Games" },
    { href: "/operations/annotators", label: "Annotators" },
    { href: "/operations/populations", label: "Populations" },
    { href: "/operations/runs", label: "Runs" },
    { href: "/operations/scrapes", label: "Scrapes" },
    { href: "/operations/annotations", label: "Annotations" },
    { href: "/operations/adjudication-sample", label: "Adj. sample" },
];

export function OperationsNav() {
    const pathname = usePathname();

    return (
        <nav className="flex flex-wrap gap-1.5">
            {tabs.map((tab) => {
                // Browse marks active for the whole subtree (its list and detail
                // routes); the write forms match their exact path.
                const active =
                    tab.href === "/operations/browse"
                        ? pathname.startsWith("/operations/browse")
                        : pathname === tab.href;
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