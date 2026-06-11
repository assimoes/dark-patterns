"use client";

import { useListAnnotators } from "@/hooks/useListAnnotators";


export function AnnotatorMultiSelect({
    value,
    onChange,
}: {
    value: number[];
    onChange: (ids: number[]) => void;
}) {
    const { data, isPending, isError } = useListAnnotators();
    const llms = (data ?? []).filter((a) => a.kind === "llm");

    const toggle = (id: number) => {
        onChange(
            value.includes(id) ? value.filter((x) => x !== id) : [...value, id],
        );
    };

    return (
        <div className="flex flex-col gap-1.5">
            <span className="text-xs font-medium text-slate-600">Annotators</span>

            {isPending ? (
                <p className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-400">
                    Loading annotators…
                </p>
            ) : isError ? (
                <p className="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs text-rose-700">
                    Couldn&apos;t load annotators.
                </p>
            ) : llms.length === 0 ? (
                <p className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-400">
                    No llm annotators available — add one first.
                </p>
            ) : (
                <div className="flex flex-col gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2.5">
                    {llms.map((a) => (
                        <label
                            key={a.id}
                            className="flex items-center gap-2.5 text-sm text-slate-700"
                        >
                            <input
                                type="checkbox"
                                checked={value.includes(a.id)}
                                onChange={() => toggle(a.id)}
                                className="size-4 rounded border-slate-300 text-slate-900 focus:ring-slate-900/10"
                            />
                            <span className="font-medium">{a.label}</span>
                            {a.model ? (
                                <span className="font-mono text-xs text-slate-400">
                                    {a.model}
                                </span>
                            ) : null}
                        </label>
                    ))}
                </div>
            )}

            <span className="text-xs text-slate-400">
                Tick the llm members that form the run&apos;s panel.
            </span>
        </div>
    );
}