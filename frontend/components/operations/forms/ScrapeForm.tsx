"use client";

import { useMemo, useState } from "react";
import { ErrorPanel, Field, SubmitButton, SuccessPanel, inputClass } from "@/components/operations/form";
import { GameSelect } from "@/components/operations/GameSelect";
import { useEnqueueScrape } from "@/hooks/useEnqueueScrape";
import { useListGames } from "@/hooks/useListGames";
import { ApiError } from "@/lib/api/utils";
import type { EnqueueScrapeInput, ScrapeFilter } from "@/lib/types";

// ScrapeForm enqueues a scrape for a game on one source. the source options come from the game's own
// source_refs, so adding a new source to a game makes it scrapeable here with no code change. the
// backend resolves the target handle from those refs.
export function ScrapeForm({ presetGameId }: { presetGameId?: string }) {
    const [game, setGame] = useState(presetGameId ?? "");
    const [source, setSource] = useState("");
    const [filter, setFilter] = useState<"" | ScrapeFilter>("");
    const [lang, setLang] = useState("");
    const [max, setMax] = useState("");

    const games = useListGames();
    const selected = games.data?.find((g) => g.id === game);
    const sources = useMemo(() => Object.keys(selected?.sourceRefs ?? {}), [selected]);

    const enqueueScrape = useEnqueueScrape();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: EnqueueScrapeInput = { game_id: Number(game), source };
        if (source === "steam" && filter) body.filter = filter;
        if (lang.trim()) body.lang = lang.trim();
        if (max !== "") body.max = Number(max);
        enqueueScrape.mutate(body);
    };

    const errorMessage =
        enqueueScrape.error instanceof ApiError
            ? enqueueScrape.error.message || "Request failed."
            : enqueueScrape.isError
                ? "Could not enqueue the scrape."
                : null;

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <GameSelect value={game} onChange={(v) => { setGame(v); setSource(""); }} required />

            <Field
                label="Source"
                hint={
                    !game
                        ? "Pick a game first."
                        : sources.length === 0
                            ? "This game has no source references. Add one on the game."
                            : "The game's registered sources. The handle is resolved from the game."
                }
            >
                <select
                    required
                    value={source}
                    onChange={(e) => setSource(e.target.value)}
                    className={inputClass}
                    disabled={!game || sources.length === 0}
                >
                    <option value="">— select a source —</option>
                    {sources.map((s) => (
                        <option key={s} value={s}>
                            {s} ({selected?.sourceRefs[s]})
                        </option>
                    ))}
                </select>
            </Field>

            {source === "steam" ? (
                <Field label="Filter">
                    <select value={filter} onChange={(e) => setFilter(e.target.value as "" | ScrapeFilter)} className={inputClass}>
                        <option value="">— default —</option>
                        <option value="recent">recent</option>
                        <option value="updated">updated</option>
                    </select>
                </Field>
            ) : null}

            <Field label="Language" hint="Optional. Used by sources that filter on it (e.g. steam).">
                <input type="text" value={lang} onChange={(e) => setLang(e.target.value)} className={inputClass} />
            </Field>

            <Field label="Max items">
                <input type="number" min="0" value={max} onChange={(e) => setMax(e.target.value)} className={inputClass} />
            </Field>

            <ErrorPanel message={errorMessage} />

            <SubmitButton pending={enqueueScrape.isPending} idleLabel="Enqueue scrape" pendingLabel="Enqueuing…" />

            {enqueueScrape.data ? <SuccessPanel title="Scrape enqueued" data={enqueueScrape.data} /> : null}
        </form>
    );
}
