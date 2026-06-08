"use client";

import { useState } from 'react'
import { useRouter } from 'next/navigation';

import { api } from '@/lib/api';
import type { DecisionInput } from '@/types/adjudication';


export function useDecision() {
    const router = useRouter();
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    async function decide(input: DecisionInput) {
        setPending(true);
        setError(null);

        try {
            await api.decide(input);
            router.push(`/?gold_run=${input.gold_run}&panel_run=${input.panel_run}`);
        } catch (e) {
            setError(e instanceof Error ? e.message : "failed");
        } finally {
            setPending(false);
        }
    }

    return { decide, pending, error }
}