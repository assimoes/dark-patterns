import type { ReactNode } from 'react';

export function Badge({ tone, children }: { tone: "present" | "absent"; children: ReactNode }) {
    const cls = tone === "present" ? "bg-red-100 text-red-800" : "bg-gray-100 text-gray-600";

    return <span className={`inline-block rounded px-2 py-0.5 text-xs font-medium ${cls}`}>
        {children}
    </span>
}