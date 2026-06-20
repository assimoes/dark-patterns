"use client";

import { useEffect, type ReactNode } from "react";
import { X } from "lucide-react";

// Drawer is a right-hand slide-over used for create and edit forms. Esc and the backdrop close it.
export function Drawer({
    open,
    onClose,
    title,
    subtitle,
    children,
}: {
    open: boolean;
    onClose: () => void;
    title: string;
    subtitle?: string;
    children: ReactNode;
}) {
    useEffect(() => {
        if (!open) return;
        const onKey = (e: KeyboardEvent) => {
            if (e.key === "Escape") onClose();
        };
        window.addEventListener("keydown", onKey);
        return () => window.removeEventListener("keydown", onKey);
    }, [open, onClose]);

    if (!open) return null;

    return (
        <div className="fixed inset-0 z-50 flex justify-end">
            <div
                className="absolute inset-0 bg-slate-900/30 backdrop-blur-sm"
                onClick={onClose}
            />
            <div className="relative flex h-full w-full max-w-md flex-col overflow-y-auto border-l border-slate-200 bg-white shadow-xl">
                <header className="flex items-start justify-between gap-4 border-b border-slate-100 px-6 py-4">
                    <div>
                        <h2 className="text-sm font-semibold text-slate-900">{title}</h2>
                        {subtitle ? (
                            <p className="text-xs text-slate-500">{subtitle}</p>
                        ) : null}
                    </div>
                    <button
                        type="button"
                        onClick={onClose}
                        className="grid size-7 shrink-0 place-items-center rounded-lg text-slate-400 transition-colors hover:bg-slate-50 hover:text-slate-700"
                        aria-label="Close"
                    >
                        <X className="size-4" />
                    </button>
                </header>
                <div className="px-6 py-5">{children}</div>
            </div>
        </div>
    );
}
