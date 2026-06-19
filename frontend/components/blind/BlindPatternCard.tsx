import { Check, X, ChevronDown, Highlighter } from "lucide-react";
import { families, type Pattern } from "@/lib/codebook";
import { Decision } from "@/lib/types";

// State here is purely the labeller's own decision — there is no panel signal.
const STATE = {
    undecided: { bar: "bg-slate-200", tint: "bg-white", chip: "bg-slate-100 text-slate-500", label: "not decided" },
    present: { bar: "bg-emerald-400", tint: "bg-emerald-50/50", chip: "bg-emerald-100 text-emerald-700", label: "present" },
    absent: { bar: "bg-slate-400", tint: "bg-white", chip: "bg-slate-100 text-slate-600", label: "absent" },
} as const;

export function BlindPatternCard({
    pattern,
    decision,
    evidence,
    capturing,
    canCite = true,
    evidenceColor,
    onDecide,
    onCite,
    onCancelCite,
    onClearEvidence,
    onHover,
}: {
    pattern: Pattern;
    decision?: Decision;
    evidence?: string;
    capturing: boolean;
    canCite?: boolean;
    evidenceColor?: string;
    onDecide: (d: Decision) => void;
    onCite: () => void;
    onCancelCite: () => void;
    onClearEvidence: () => void;
    onHover: (code: string | null) => void;
}) {
    const state = STATE[decision ?? "undecided"];

    return (
        <div
            onMouseEnter={() => onHover(pattern.code)}
            onMouseLeave={() => onHover(null)}
            className={`flex overflow-hidden rounded-xl border ${capturing ? "border-sky-300 ring-2 ring-sky-200" : "border-slate-200"
                }`}
        >
            <div className={`w-1.5 shrink-0 ${state.bar}`} />

            <div className={`min-w-0 flex-1 px-4 py-3.5 ${state.tint}`}>
                <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                        <div className="flex items-center gap-2">
                            <span className="font-mono text-xs font-semibold text-slate-500">{pattern.code}</span>
                            <h3 className="truncate text-sm font-semibold text-slate-900">{pattern.name}</h3>
                        </div>
                        <p className="text-xs text-slate-400">{families[pattern.family]}</p>
                    </div>
                    <span className={`shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium ${state.chip}`}>
                        {state.label}
                    </span>
                </div>

                <p className="mt-3 text-xs leading-5 text-slate-600">{pattern.definition}</p>

                <details className="group mt-2">
                    <summary className="flex cursor-pointer list-none items-center gap-1 text-xs font-medium text-slate-500 hover:text-slate-700">
                        <ChevronDown className="size-3.5 transition-transform group-open:rotate-180" />
                        Examples &amp; counter-examples
                    </summary>
                    <div className="mt-2 space-y-2 border-l-2 border-slate-200 pl-3">
                        {pattern.examples.map((ex, i) => (
                            <p key={`e${i}`} className="flex gap-1.5 text-xs text-slate-600">
                                <Check className="mt-0.5 size-3 shrink-0 text-emerald-500" />
                                <span>{ex}</span>
                            </p>
                        ))}
                        {pattern.counterExamples.map((ex, i) => (
                            <p key={`c${i}`} className="flex gap-1.5 text-xs text-slate-500">
                                <X className="mt-0.5 size-3 shrink-0 text-rose-400" />
                                <span>{ex}</span>
                            </p>
                        ))}
                    </div>
                </details>

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

                    {decision === "present" && canCite ? (
                        evidence ? null : capturing ? (
                            <button
                                type="button"
                                onClick={onCancelCite}
                                className="ml-auto text-xs font-medium text-sky-600 underline-offset-2 hover:underline"
                            >
                                selecting… cancel
                            </button>
                        ) : (
                            <button
                                type="button"
                                onClick={onCite}
                                className="ml-auto inline-flex items-center gap-1 rounded-md border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50"
                            >
                                <Highlighter className="size-3" /> Cite evidence
                            </button>
                        )
                    ) : null}
                </div>

                {/* the evidence the labeller cited */}
                {decision === "present" && evidence ? (
                    <div className="mt-2 flex items-start gap-2">
                        <p
                            className="flex-1 border-l-2 pl-2 text-xs italic text-slate-500"
                            style={{ borderColor: evidenceColor ?? "#cbd5e1" }}
                        >
                            &ldquo;{evidence}&rdquo;
                        </p>
                        <button
                            type="button"
                            onClick={onClearEvidence}
                            className="shrink-0 rounded p-0.5 text-slate-300 transition-colors hover:text-slate-600"
                            aria-label="Remove evidence"
                        >
                            <X className="size-3.5" />
                        </button>
                    </div>
                ) : null}
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