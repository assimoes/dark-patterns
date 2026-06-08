"use client";

import { Button } from "@/components/ui/Button";
import { useDecision } from "../hooks/useDecision";

export function DecisionForm(props: {
    goldRun: number;
    panelRun: number;
    individualId: number;
    patternId: number;
}) {
    const { decide, pending, error } = useDecision();
    const base = {
        gold_run: props.goldRun,
        panel_run: props.panelRun,
        individual_id: props.individualId,
        pattern_id: props.patternId,
    };

    return (
        <section aria-label="your decision" className="flex flex-col gap-3">
            <div className="flex gap-3">
                <Button disabled={pending} onClick={() => decide({ ...base, label: true })}>
                    present
                </Button>
                <Button disabled={pending} variant="secondary" onClick={() => decide({ ...base, label: false })}>
                    absent
                </Button>
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
        </section>
    );
}