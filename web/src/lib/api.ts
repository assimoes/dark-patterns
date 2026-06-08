import type {
    Cell,
    DecisionInput,
    Review,
    ReviewItem,
    ReviewDecisionInput,
} from "@/types/adjudication";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function getJSON<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${BASE}${path}`, {
        ...init,
        headers: { "Content-Type": "application/json", ...init?.headers },
        cache: "no-store",
    });

    if (!res.ok) throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);

    return res.json() as Promise<T>;
}

export const api = {
    worklist: (p: { goldRun: number; panelRun: number; perGame: number; taxVersion: number }) =>
        getJSON<ReviewItem[]>(
            `/api/worklist?gold_run=${p.goldRun}&panel_run=${p.panelRun}&per_game=${p.perGame}&tax_version=${p.taxVersion}`,
        ),
    review: (individual: number, p: { panelRun: number; goldRun: number; taxVersion: number }) =>
        getJSON<Review>(
            `/api/reviews/${individual}?panel_run=${p.panelRun}&gold_run=${p.goldRun}&tax_version=${p.taxVersion}`,
        ),
    decideReview: (individual: number, input: ReviewDecisionInput) =>
        getJSON<{ status: string; saved: number }>(`/api/reviews/${individual}/decisions`, {
            method: "POST",
            body: JSON.stringify(input),
        }),
    cell: (individual: number, pattern: number, panelRun: number) =>
        getJSON<Cell>(`/api/cells/${individual}/${pattern}?panel_run=${panelRun}`),
    decide: (input: DecisionInput) =>
        getJSON<{ status: string; direction: string }>(`/api/decisions`, {
            method: "POST",
            body: JSON.stringify(input),
        }),
};

