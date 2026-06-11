"use client";

import { useState } from "react";
import { ListChecks } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    SubmitButton,
    SuccessPanel,
} from "@/components/operations/form";
import { RunSelect } from "@/components/operations/RunSelect";
import { useEnqueueAnnotations } from "@/hooks/useEnqueueAnnotations";
import { ApiError } from "@/lib/api";

// Enqueue-annotations form.
export default function AnnotationsPage() {
    const [runId, setRunId] = useState("");

    const enqueueAnnotations = useEnqueueAnnotations();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        enqueueAnnotations.mutate(runId.trim());
    };

    const errorMessage =
        enqueueAnnotations.error instanceof ApiError
            ? enqueueAnnotations.error.message || "Request failed."
            : enqueueAnnotations.isError
                ? "Could not enqueue annotations. Please try again."
                : null;

    return (
        <Card
            icon={<ListChecks className="size-5" />}
            iconClass="bg-slate-100 text-slate-700"
            title="Enqueue annotations"
            subtitle="Kick off annotation jobs for an existing run."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
                <RunSelect
                    value={runId}
                    onChange={setRunId}
                    runTypeFilter="llm_panel"
                    label="Run"
                    required
                />

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={enqueueAnnotations.isPending}
                    idleLabel="Enqueue annotations"
                    pendingLabel="Enqueuing…"
                />

                {enqueueAnnotations.data ? (
                    <SuccessPanel
                        title="Annotations enqueued"
                        data={enqueueAnnotations.data}
                    />
                ) : null}
            </form>
        </Card>
    );
}