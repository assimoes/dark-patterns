"use client";

import { type ReactNode } from "react";

// Panel shell
export function BrowsePanel({
    title,
    subtitle,
    badge,
    children,
}: {
    title: string;
    subtitle?: string;
    badge?: ReactNode;
    children: ReactNode;
}) {
    return (
        <section className="flex flex-col rounded-2xl border border-slate-200/80 bg-white/90 shadow-sm ring-1 ring-slate-900/[0.02] backdrop-blur-sm">
            <header className="flex items-start justify-between gap-4 border-b border-slate-100 px-6 py-5">
                <div>
                    <h2 className="text-sm font-semibold tracking-tight text-slate-900">
                        {title}
                    </h2>
                    {subtitle ? (
                        <p className="text-xs text-slate-500">{subtitle}</p>
                    ) : null}
                </div>
                {badge}
            </header>
            <div className="flex-1 overflow-x-auto px-6 py-5">{children}</div>
        </section>
    );
}

// Table header row from a list of column labels.
export function TableHead({ columns }: { columns: string[] }) {
    return (
        <thead>
            <tr className="border-b border-slate-100 text-left text-xs font-medium text-slate-400">
                {columns.map((c) => (
                    <th key={c} className="px-3 py-2 font-medium">
                        {c}
                    </th>
                ))}
            </tr>
        </thead>
    );
}

// Loading state
export function TableSkeleton({
    columns,
    rows = 5,
}: {
    columns: number;
    rows?: number;
}) {
    return (
        <div className="animate-pulse space-y-2">
            {Array.from({ length: rows }).map((_, r) => (
                <div key={r} className="flex gap-3">
                    {Array.from({ length: columns }).map((__, c) => (
                        <div key={c} className="h-4 flex-1 rounded bg-slate-100" />
                    ))}
                </div>
            ))}
        </div>
    );
}

// Error state
export function TableError({ onRetry }: { onRetry: () => void }) {
    return (
        <div className="py-6 text-center">
            <p className="text-sm text-slate-500">
                The API request failed. Check that the backend is running.
            </p>
            <button
                type="button"
                onClick={onRetry}
                className="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700"
            >
                Retry
            </button>
        </div>
    );
}

// Empty state
export function TableEmpty({ message }: { message: string }) {
    return <p className="py-6 text-center text-sm text-slate-400">{message}</p>;
}