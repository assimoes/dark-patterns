import { Check, X, ChevronDown } from "lucide-react";
import { families } from "@/lib/codebook";
import { isOverride, type PatternVote } from "@/lib/adjudication";
import { Decision } from "@/lib/types";

// Fixed colour states (literal classes so Tailwind keeps them).
const STATE = {
    green: { bar: "bg-emerald-400", tint: "bg-emerald-50/50", chip: "bg-emerald-100 text-emerald-700", label: "majority detected" },
    yellow: { bar: "bg-amber-400", tint: "bg-amber-50/50", chip: "bg-amber-100 text-amber-700", label: "minority / tie" },
    gray: { bar: "bg-slate-300", tint: "bg-white", chip: "bg-slate-100 text-slate-600", label: "none detected" },
    blue: { bar: "bg-sky-400", tint: "bg-sky-50/60", chip: "bg-sky-100 text-sky-700", label: "override" },
} as const;

export function PatternCard({
    pv,
    decision,
    evidenceColor,
    onDecide,
    onHover,
}: {
    pv: PatternVote;
    decision?: Decision;
    evidenceColor?: string;
    onDecide: (d: Decision) => void;
    onHover: (code: string | null) => void;
}) {
    const override = isOverride(pv.verdict, decision);
    const state = STATE[override ? "blue" : pv.signal];

    return (
        <div
            onMouseEnter={() => onHover(pv.pattern.code)}
            onMouseLeave={() => onHover(null)}
            className="flex overflow-hidden rounded-xl border border-slate-200"
        >
            <div className={`w-1.5 shrink-0 ${state.bar}`} />

            <div className={`min-w-0 flex-1 px-4 py-3.5 ${state.tint}`}>
                {/* header */}
                <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                        <div className="flex items-center gap-2">
                            <span className="font-mono text-xs font-semibold text-slate-500">{pv.pattern.code}</span>
                            <h3 className="truncate text-sm font-semibold text-slate-900">{pv.pattern.name}</h3>
                        </div>
                        <p className="text-xs text-slate-400">{families[pv.pattern.family]}</p>
                    </div>
                    <span className={`shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium ${state.chip}`}>
                        {state.label}
                    </span>
                </div>

                {/* per-model votes */}
                <div className="mt-3 flex flex-wrap gap-1.5">
                    {pv.votes.map((v) => (
                        <span
                            key={v.modelId}
                            className={`inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 text-[11px] font-medium ${v.present
                                ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                                : "border-slate-200 bg-white text-slate-400"
                                }`}
                        >
                            {v.present ? <Check className="size-3" /> : <X className="size-3" />}
                            {v.short}
                        </span>
                    ))}
                    <span className="ml-auto self-center text-xs text-slate-400 tabular-nums">
                        {pv.presentCount}/4
                    </span>
                </div>

                {/* definition + codebook disclosure */}
                <p className="mt-3 text-xs leading-5 text-slate-600">{pv.pattern.definition}</p>

                <details className="group mt-2">
                    <summary className="flex cursor-pointer list-none items-center gap-1 text-xs font-medium text-slate-500 hover:text-slate-700">
                        <ChevronDown className="size-3.5 transition-transform group-open:rotate-180" />
                        Examples &amp; counter-examples
                    </summary>
                    <div className="mt-2 space-y-2 border-l-2 border-slate-200 pl-3">
                        {pv.pattern.examples.map((ex, i) => (
                            <p key={`e${i}`} className="flex gap-1.5 text-xs text-slate-600">
                                <Check className="mt-0.5 size-3 shrink-0 text-emerald-500" />
                                <span>{ex}</span>
                            </p>
                        ))}
                        {pv.pattern.counterExamples.map((ex, i) => (
                            <p key={`c${i}`} className="flex gap-1.5 text-xs text-slate-500">
                                <X className="mt-0.5 size-3 shrink-0 text-rose-400" />
                                <span>{ex}</span>
                            </p>
                        ))}
                    </div>
                </details>

                {/* cited evidence */}
                {pv.evidence.length > 0 ? (
                    <div className="mt-3 space-y-1">
                        {pv.evidence.map((q, i) => (
                            <p
                                key={i}
                                className="border-l-2 pl-2 text-xs italic text-slate-500"
                                style={{ borderColor: evidenceColor ?? "#cbd5e1" }}
                            >
                                &ldquo;{q}&rdquo;
                            </p>
                        ))}
                    </div>
                ) : null}

                {/* decision */}
                <div className="mt-3.5 flex flex-wrap items-center gap-3 border-t border-slate-100 pt-3">
                    <div className="inline-flex rounded-lg border border-slate-200 bg-white p-0.5">
                        <DecisionButton active={decision === "present"} tone="present" onClick={() => onDecide("present")}>
                            Present
                        </DecisionButton>
                        <DecisionButton active={decision === "absent"} tone="absent" onClick={() => onDecide("absent")}>
                            Absent
                        </DecisionButton>
                    </div>

                    <StatusChip verdict={pv.verdict} decision={decision} override={override} />

                    {pv.verdict !== "tie" ? (
                        override ? (
                            <button
                                type="button"
                                onClick={() => onDecide(pv.verdict as Decision)}
                                className="ml-auto text-xs font-medium text-slate-500 underline-offset-2 hover:text-slate-700 hover:underline"
                            >
                                Agree with panel
                            </button>
                        ) : (
                            <button
                                type="button"
                                onClick={() => onDecide(pv.verdict === "present" ? "absent" : "present")}
                                className="ml-auto rounded-md border border-sky-200 bg-white px-2.5 py-1 text-xs font-medium text-sky-700 transition-colors hover:bg-sky-50"
                            >
                                Disagree
                            </button>
                        )
                    ) : null}
                </div>
            </div>
        </div>
    );
}

function DecisionButton({
    active,
    tone,
    onClick,
    children,
}: {
    active: boolean;
    tone: "present" | "absent";
    onClick: () => void;
    children: React.ReactNode;
}) {
    const activeCls =
        tone === "present" ? "bg-emerald-600 text-white shadow-sm" : "bg-slate-700 text-white shadow-sm";
    return (
        <button
            type="button"
            onClick={onClick}
            className={`rounded-md px-3 py-1 text-xs font-semibold transition-colors ${active ? activeCls : "text-slate-500 hover:text-slate-800"
                }`}
        >
            {children}
        </button>
    );
}

function StatusChip({
    verdict,
    decision,
    override,
}: {
    verdict: PatternVote["verdict"];
    decision?: Decision;
    override: boolean;
}) {
    if (!decision) {
        return <span className="text-xs text-slate-400">not decided</span>;
    }
    if (override) {
        return <span className="text-xs font-medium text-sky-600">overrides the panel</span>;
    }
    if (verdict === "tie") {
        return <span className="text-xs font-medium text-amber-600">your call (tie)</span>;
    }
    return <span className="text-xs font-medium text-slate-500">matches the panel</span>;
}