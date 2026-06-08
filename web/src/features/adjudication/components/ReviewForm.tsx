"use client";

import { useState } from 'react';
import { useRouter } from 'next/navigation';

import { api } from '@/lib/api';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import type { Review, ReviewPattern } from '@/types/adjudication';

export function ReviewForm({
    review,
    goldRun,
    panelRun,
    backHref,
}: {
    review: Review;
    goldRun: number;
    panelRun: number;
    backHref: string;
}) {
    const router = useRouter();
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const [checked, setChecked] = useState<Record<number, boolean>>(() => {
        const init: Record<number, boolean> = {};
        for (const p of review.patterns) {
            init[p.pattern_id] = p.decided ? p.final_label : p.panel_vote;
        }
        return init;
    });

    function toggle(id: number) {
        setChecked((c) => ({ ...c, [id]: !c[id] }));
    }

    async function save() {
        setPending(true);
        setError(null);
        try {
            await api.decideReview(review.individual_id, {
                gold_run: goldRun,
                panel_run: panelRun,
                decisions: review.patterns.map((p) => ({
                    pattern_id: p.pattern_id,
                    label: !!checked[p.pattern_id],
                })),
            });
            router.push(backHref);
        } catch (e) {
            setError(e instanceof Error ? e.message : 'failed');
        } finally {
            setPending(false);
        }
    }

    const families: Record<string, ReviewPattern[]> = {};
    for (const p of review.patterns) {
        (families[p.family] ??= []).push(p);
    }

    return (
        <div className="space-y-6">
            {Object.entries(families).map(([family, patterns]) => (
                <section key={family} className="space-y-2">
                    <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-500">{family}</h2>
                    <ul className="space-y-2">
                        {patterns.map((p) => (
                            <li key={p.pattern_id} className="rounded-lg border p-3">
                                <label className="flex cursor-pointer items-start gap-3">
                                    <input
                                        type="checkbox"
                                        className="mt-1 h-4 w-4"
                                        checked={!!checked[p.pattern_id]}
                                        onChange={() => toggle(p.pattern_id)}
                                    />
                                    <div className="flex-1">
                                        <div className="flex items-center gap-2">
                                            <span className="font-medium">{p.code} - {p.name}</span>
                                            {p.n_present > 0 && (
                                                <Badge tone="present">{p.n_present}/{p.n_total} flagged</Badge>
                                            )}
                                        </div>
                                        <p className="text-sm text-gray-600">{p.description}</p>

                                        {p.detections.length > 0 && (
                                            <ul className="mt-2 space-y-1 border-l-2 border-gray-200 pl-3">
                                                {p.detections.map((d) => (
                                                    <li key={d.model_slug} className="text-xs">
                                                        <span className="font-mono text-gray-700">{d.model_slug}</span>
                                                        {d.evidence && (
                                                            <span className="text-gray-600"> - <span className="italic">&quot;{d.evidence}&quot;</span></span>
                                                        )}
                                                        {d.explanation && (
                                                            <p className="text-gray-500">{d.explanation}</p>
                                                        )}
                                                    </li>
                                                ))}
                                            </ul>
                                        )}
                                    </div>
                                </label>
                            </li>
                        ))}
                    </ul>
                </section>
            ))}

            <div className="sticky bottom-0 flex items-center gap-3 border-t bg-white py-3">
                <Button disabled={pending} onClick={save}>
                    {pending ? 'saving' : 'save review'}
                </Button>
                {error && <p className="text-sm text-red-600">{error}</p>}
            </div>
        </div>
    );
}
