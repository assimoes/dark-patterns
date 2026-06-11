import { type ReactNode } from "react";

export function Card({
    icon,
    iconClass = "bg-slate-100 text-slate-700",
    title,
    subtitle,
    badge,
    children,
    className = "",
}: {
    icon: ReactNode;
    iconClass?: string;
    title: string;
    subtitle?: string;
    badge?: ReactNode;
    children: ReactNode;
    className?: string;
}) {
    return (
        <section
            className={`flex flex-col rounded-2xl border border-slate-200/80 bg-white/90 shadow-sm ring-1 ring-slate-900/[0.02] backdrop-blur-sm ${className}`}
        >
            <header className="flex items-start justify-between gap-4 border-b border-slate-100 px-6 py-5">
                <div className="flex items-center gap-3">
                    <span className={`grid size-10 shrink-0 place-items-center rounded-xl ${iconClass}`}>
                        {icon}
                    </span>
                    <div>
                        <h2 className="text-sm font-semibold tracking-tight text-slate-900">{title}</h2>
                        {subtitle ? <p className="text-xs text-slate-500">{subtitle}</p> : null}
                    </div>
                </div>
                {badge}
            </header>
            <div className="flex-1 px-6 py-5">{children}</div>
        </section>
    );
}

export function Badge({
    children,
    className = "",
}: {
    children: ReactNode;
    className?: string;
}) {
    return (
        <span
            className={`inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2.5 py-1 text-xs font-medium text-slate-600 ${className}`}
        >
            {children}
        </span>
    );
}