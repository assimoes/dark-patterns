import type { Vote } from "@/types/adjudication";
import { Badge } from '@/components/ui/Badge'


export function PanelVotes({
    vote,
    nPresent,
    nTotal,
    votes,
}: {
    vote: boolean;
    nPresent: number;
    nTotal: number;
    votes: Vote[];
}) {
    return (
        <section aria-label="panel votes" className="rounded-lg border p-4">
            <div className="mb-3 flex items-center gap-2">
                <span className="font-medium">panel majority:</span>
                <Badge tone={vote ? "present" : "absent"}>
                    {vote ? "present" : "absent"}
                </Badge>
                <span className="text-sm text-gray-500">
                    ({nPresent} / {nTotal} present)
                </span>
            </div>

            <ul className="space-y-2">
                {votes.map((v) => (
                    <li key={v.model_slug} className="rounded border p-2">
                        <div className="flex justify-between text-sm">
                            <span className="font-mono">{v.model_slug}</span>
                            <Badge tone={v.present ? "present" : "absent"}>{v.present ? "present" : "absent"}</Badge>
                        </div>
                        {v.evidence && (
                            <p className="mt-1 text-xs text-gray-700">
                                <span className="font-semibold">evidence:</span> {" "}
                                <span className="italic">"{v.evidence}"</span>
                            </p>
                        )}
                        {v.explanation && (
                            <p className="mt-1 text-xs text-gray-600">
                                <span className="font-semibold">explanation:</span> {v.explanation}
                            </p>
                        )}
                    </li>
                ))}
            </ul>
        </section>
    );
}