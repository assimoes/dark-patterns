"use client";

import { useState } from "react";
import { Download } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    Field,
    SubmitButton,
    SuccessPanel,
    inputClass,
} from "@/components/operations/form";
import { useEnqueueScrape } from "@/hooks/useEnqueueScrape";
import { ApiError } from "@/lib/api/utils";
import type { EnqueueScrapeInput, ScrapeFilter, ScrapeSource } from "@/lib/types";
import { GameSelect } from "@/components/operations/GameSelect";

export default function ScrapesPage() {
    const [game, setGame] = useState("");
    const [source, setSource] = useState<ScrapeSource>("steam");
    const [target, setTarget] = useState("");
    const [filter, setFilter] = useState<"" | ScrapeFilter>("");
    const [lang, setLang] = useState("");
    const [max, setMax] = useState("");

    const enqueueScrape = useEnqueueScrape();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        // steam's app id is the game id, so a blank target falls back to it. reddit needs the subreddit.
        const resolved = target.trim() || (source === "steam" ? game : "");
        const body: EnqueueScrapeInput = {
            game_id: Number(game),
            source,
            target: resolved,
        };
        if (source === "steam" && filter) body.filter = filter;
        if (source === "steam" && lang.trim()) body.lang = lang.trim();
        if (max !== "") body.max = Number(max);
        enqueueScrape.mutate(body);
    };

    const errorMessage =
        enqueueScrape.error instanceof ApiError
            ? enqueueScrape.error.message || "Request failed."
            : enqueueScrape.isError
                ? "Could not enqueue the scrape. Please try again."
                : null;

    return (
        <Card
            icon={<Download className="size-5" />}
            iconClass="bg-rose-50 text-rose-600"
            title="Enqueue scrape"
            subtitle="Queue a scrape for a game on one of its sources."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
                <GameSelect value={game} onChange={setGame} required />

                <Field label="Source">
                    <select
                        value={source}
                        onChange={(e) => setSource(e.target.value as ScrapeSource)}
                        className={inputClass}
                    >
                        <option value="steam">steam</option>
                        <option value="reddit">reddit</option>
                    </select>
                </Field>

                <Field
                    label={source === "reddit" ? "Subreddit" : "Steam app id"}
                    hint={source === "reddit" ? "e.g. r/EVE." : "Defaults to the game id."}
                >
                    <input
                        type="text"
                        value={target}
                        onChange={(e) => setTarget(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                {source === "steam" ? (
                    <>
                        <Field label="Filter">
                            <select
                                value={filter}
                                onChange={(e) => setFilter(e.target.value as "" | ScrapeFilter)}
                                className={inputClass}
                            >
                                <option value="">— default —</option>
                                <option value="recent">recent</option>
                                <option value="updated">updated</option>
                            </select>
                        </Field>

                        <Field label="Language" hint="e.g. english.">
                            <input
                                type="text"
                                value={lang}
                                onChange={(e) => setLang(e.target.value)}
                                className={inputClass}
                            />
                        </Field>
                    </>
                ) : null}

                <Field label="Max items">
                    <input
                        type="number"
                        min="0"
                        value={max}
                        onChange={(e) => setMax(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={enqueueScrape.isPending}
                    idleLabel="Enqueue scrape"
                    pendingLabel="Enqueuing…"
                />

                {enqueueScrape.data ? (
                    <SuccessPanel title="Scrape enqueued" data={enqueueScrape.data} />
                ) : null}
            </form>
        </Card>
    );
}
