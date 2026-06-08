import { api } from "@/lib/api";
import { ReviewPanel } from "@/features/adjudication/components/ReviewPanel";
import { PanelVotes } from "@/features/adjudication/components/PanelVotes";
import { DecisionForm } from "@/features/adjudication/components/DecisionForm";


export default async function CellPage({
    params,
    searchParams,
}: {
    params: Promise<{ individual: string; pattern: string; }>;
    searchParams: Promise<{ panel_run?: string; gold_run?: string }>;
}) {
    const { individual, pattern } = await params;
    const { panel_run = "0", gold_run = "0" } = await searchParams;

    const cell = await api.cell(Number(individual), Number(pattern), Number(panel_run));

    return (
        <main className="mx-auto max-w-3xl space-y-6 p-6">
            <h1 className="text-xl font-semibold">
                {cell.code} - {cell.name}
            </h1>
            <ReviewPanel text={cell.review_text} />
            <PanelVotes vote={cell.panel_vote} nPresent={cell.n_present} nTotal={cell.n_total} votes={cell.votes} />
            <DecisionForm
                goldRun={Number(gold_run)}
                panelRun={Number(panel_run)}
                individualId={cell.individual_id}
                patternId={cell.pattern_id}
            />
        </main>
    );
}

