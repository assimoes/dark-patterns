"use client";

import { type ReactNode } from "react";
import { AlertTriangle } from "lucide-react";

// ConfirmDialog is a centered modal for destructive confirms. items lists the blast radius of a
// cascade delete (label + count); leave it empty for a plain confirm.
export function ConfirmDialog({
    open,
    title,
    body,
    items,
    confirmLabel,
    pendingLabel = "Deleting…",
    pending,
    error,
    onConfirm,
    onCancel,
}: {
    open: boolean;
    title: string;
    body?: ReactNode;
    items?: { label: string; value: number }[];
    confirmLabel: string;
    pendingLabel?: string;
    pending?: boolean;
    error?: string | null;
    onConfirm: () => void;
    onCancel: () => void;
}) {
    if (!open) return null;

    return (
        <div className="fixed inset-0 z-50 grid place-items-center p-4">
            <div className="absolute inset-0 bg-slate-900/30 backdrop-blur-sm" onClick={onCancel} />
            <div className="relative w-full max-w-sm rounded-2xl border border-slate-200 bg-white p-5 shadow-xl">
                <div className="flex items-start gap-3">
                    <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-rose-50 text-rose-600">
                        <AlertTriangle className="size-4.5" />
                    </span>
                    <div className="min-w-0">
                        <h2 className="text-sm font-semibold text-slate-900">{title}</h2>
                        {body ? <p className="mt-1 text-sm text-slate-500">{body}</p> : null}
                    </div>
                </div>

                {items && items.length > 0 ? (
                    <ul className="mt-3 space-y-1 rounded-lg border border-slate-100 bg-slate-50/60 px-3 py-2 text-xs text-slate-600">
                        {items.map((it) => (
                            <li key={it.label} className="flex items-center justify-between gap-3">
                                <span>{it.label}</span>
                                <span className="font-mono tabular-nums text-slate-800">{it.value}</span>
                            </li>
                        ))}
                    </ul>
                ) : null}

                {error ? (
                    <p className="mt-3 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-medium text-rose-700">
                        {error}
                    </p>
                ) : null}

                <div className="mt-5 flex justify-end gap-2">
                    <button
                        type="button"
                        onClick={onCancel}
                        className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                    >
                        Cancel
                    </button>
                    <button
                        type="button"
                        onClick={onConfirm}
                        disabled={pending}
                        className="rounded-lg bg-rose-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        {pending ? pendingLabel : confirmLabel}
                    </button>
                </div>
            </div>
        </div>
    );
}
