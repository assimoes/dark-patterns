import { type ReactNode } from "react";

export const inputClass =
    "rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 outline-none transition-colors focus:border-slate-400 focus:ring-2 focus:ring-slate-900/5";

export function Field({
    label,
    hint,
    children,
}: {
    label: string;
    hint?: string;
    children: ReactNode;
}) {
    return (
        <label className="flex flex-col gap-1.5">
            <span className="text-xs font-medium text-slate-600">{label}</span>
            {children}
            {hint ? <span className="text-xs text-slate-400">{hint}</span> : null}
        </label>
    );
}

export function SubmitButton({
    pending,
    idleLabel,
    pendingLabel,
    icon,
}: {
    pending: boolean;
    idleLabel: string;
    pendingLabel: string;
    icon?: ReactNode;
}) {
    return (
        <button
            type="submit"
            disabled={pending}
            className="mt-1 inline-flex items-center justify-center gap-1.5 rounded-lg bg-slate-900 px-3.5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
        >
            {icon}
            {pending ? pendingLabel : idleLabel}
        </button>
    );
}

export function ErrorPanel({ message }: { message: string | null }) {
    if (!message) return null;
    return (
        <p className="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-medium text-rose-700">
            {message}
        </p>
    );
}

export function SuccessPanel({
    title,
    data,
}: {
    title: string;
    data: Record<string, unknown>;
}) {
    return (
        <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3">
            <p className="mb-2 text-xs font-semibold text-emerald-800">{title}</p>
            <dl className="grid grid-cols-1 gap-1.5">
                {Object.entries(data).map(([key, value]) => (
                    <div key={key} className="flex items-baseline justify-between gap-3">
                        <dt className="text-xs font-medium text-emerald-700">{key}</dt>
                        <dd className="font-mono text-xs text-emerald-900">
                            {formatValue(value)}
                        </dd>
                    </div>
                ))}
            </dl>
        </div>
    );
}

function formatValue(value: unknown): string {
    if (value === null || value === undefined) return "—";
    if (Array.isArray(value)) return value.length ? value.join(", ") : "—";
    return String(value);
}