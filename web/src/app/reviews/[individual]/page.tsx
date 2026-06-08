import { api } from '@/lib/api';
import { ReviewForm } from '@/features/adjudication/components/ReviewForm';

export default async function ReviewPage({
    params,
    searchParams,
}: {
    params: Promise<{ individual: string }>;
    searchParams: Promise<{ panel_run?: string; gold_run?: string; per_game?: string; tax_version?: string }>;
}) {
    const { individual } = await params;
    const sp = await searchParams;
    const panelRun = Number(sp.panel_run ?? 0);
    const goldRun = Number(sp.gold_run ?? 0);
    const perGame = Number(sp.per_game ?? 5);
    const taxVersion = Number(sp.tax_version ?? 1);

    const review = await api.review(Number(individual), { panelRun, goldRun, taxVersion });
    const backHref = `/?gold_run=${goldRun}&panel_run=${panelRun}&per_game=${perGame}&tax_version=${taxVersion}`;

    return (
        <main className="mx-auto max-w-3xl space-y-6 p-6">
            <header className="space-y-2">
                <h1 className="text-xl font-semibold">review {review.individual_id}</h1>
                <p className="whitespace-pre-wrap rounded-lg border bg-gray-50 p-4 text-sm leading-relaxed">
                    {review.review_text}
                </p>
            </header>
            <ReviewForm review={review} goldRun={goldRun} panelRun={panelRun} backHref={backHref} />
        </main>
    );
}
