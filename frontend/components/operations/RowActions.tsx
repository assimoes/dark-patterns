"use client";

import { type ReactNode } from "react";

export type RowAction = {
    icon: ReactNode;
    label: string;
    onClick: () => void;
    // when set, the button is disabled and the reason shows as a tooltip (the frozen case).
    disabledReason?: string;
    danger?: boolean;
};

// RowActions renders a compact row of icon buttons for a table row. A disabledReason greys the button
// out and explains why (frozen), so the operator sees the guard without clicking.
export function RowActions({ actions }: { actions: RowAction[] }) {
    return (
        <div className="flex items-center justify-end gap-1">
            {actions.map((a) => {
                const disabled = Boolean(a.disabledReason);
                const tone = a.danger
                    ? "text-rose-500 hover:bg-rose-50 hover:text-rose-700"
                    : "text-slate-400 hover:bg-slate-50 hover:text-slate-700";
                return (
                    <button
                        key={a.label}
                        type="button"
                        onClick={a.onClick}
                        disabled={disabled}
                        title={a.disabledReason ?? a.label}
                        aria-label={a.label}
                        className={`grid size-7 place-items-center rounded-lg transition-colors disabled:cursor-not-allowed disabled:text-slate-200 disabled:hover:bg-transparent ${tone}`}
                    >
                        {a.icon}
                    </button>
                );
            })}
        </div>
    );
}
