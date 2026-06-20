"use client";

import { useState } from "react";
import { Download, ImagePlus, Pencil, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/Card";
import {
    BrowsePanel,
    NewButton,
    TableEmpty,
    TableError,
    TableHead,
    TableSkeleton,
} from "@/components/operations/BrowseTable";
import { Drawer } from "@/components/operations/Drawer";
import { ConfirmDialog } from "@/components/operations/ConfirmDialog";
import { RowActions } from "@/components/operations/RowActions";
import { GameForm } from "@/components/operations/forms/GameForm";
import { ScrapeForm } from "@/components/operations/forms/ScrapeForm";
import { ImageUploadForm } from "@/components/operations/forms/ImageUploadForm";
import { useListGames } from "@/hooks/useListGames";
import { useDeleteGame } from "@/hooks/useEntityMutations";
import { ApiError } from "@/lib/api/utils";
import { fmt } from "@/lib/types";
import type { GameRow } from "@/lib/types";

type DrawerState =
    | { mode: "new" }
    | { mode: "edit"; game: GameRow }
    | { mode: "scrape"; game: GameRow }
    | { mode: "images"; game: GameRow };

export default function BrowseGamesPage() {
    const { data, isPending, isError, refetch } = useListGames();
    const games = data ?? [];

    const [drawer, setDrawer] = useState<DrawerState | null>(null);
    const [toDelete, setToDelete] = useState<GameRow | null>(null);
    const del = useDeleteGame();

    const confirmDelete = () => {
        if (!toDelete) return;
        del.mutate(toDelete.id, { onSuccess: () => setToDelete(null) });
    };

    const cancelDelete = () => {
        setToDelete(null);
        del.reset();
    };

    return (
        <>
            <BrowsePanel
                title="Games"
                subtitle="Curated games and their corpus coverage."
                badge={
                    <div className="flex items-center gap-3">
                        {data ? <Badge>{games.length} games</Badge> : null}
                        <NewButton onClick={() => setDrawer({ mode: "new" })} />
                    </div>
                }
            >
                {isPending ? (
                    <TableSkeleton columns={7} />
                ) : isError ? (
                    <TableError onRetry={() => refetch()} />
                ) : games.length === 0 ? (
                    <TableEmpty message="No games registered yet." />
                ) : (
                    <table className="w-full text-sm">
                        <TableHead columns={["Game", "Short", "Sources", "Monetization", "Reviews", "Annotated", ""]} />
                        <tbody>
                            {games.map((g) => (
                                <tr key={g.id} className="border-b border-slate-50 last:border-0">
                                    <td className="px-3 py-2.5">
                                        <span className="flex items-center gap-2">
                                            <span className="size-3 shrink-0 rounded-full" style={{ backgroundColor: g.color }} />
                                            <span className="font-medium text-slate-700">{g.name}</span>
                                        </span>
                                    </td>
                                    <td className="px-3 py-2.5 font-mono text-xs text-slate-500">{g.short}</td>
                                    <td className="px-3 py-2.5">
                                        <span className="flex flex-wrap gap-1">
                                            {Object.keys(g.sourceRefs).length === 0 ? (
                                                <span className="text-xs text-slate-300">—</span>
                                            ) : (
                                                Object.entries(g.sourceRefs).map(([k, v]) => (
                                                    <span
                                                        key={k}
                                                        title={`${k}: ${v}`}
                                                        className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-600"
                                                    >
                                                        {k}
                                                    </span>
                                                ))
                                            )}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2.5 text-slate-500">{g.monetization}</td>
                                    <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">{fmt(g.reviews)}</td>
                                    <td className="px-3 py-2.5 text-right tabular-nums text-slate-700">{fmt(g.annotated)}</td>
                                    <td className="px-3 py-2.5">
                                        <RowActions
                                            actions={[
                                                { icon: <Download className="size-4" />, label: "Scrape", onClick: () => setDrawer({ mode: "scrape", game: g }) },
                                                { icon: <ImagePlus className="size-4" />, label: "Upload images", onClick: () => setDrawer({ mode: "images", game: g }) },
                                                { icon: <Pencil className="size-4" />, label: "Edit", onClick: () => setDrawer({ mode: "edit", game: g }) },
                                                {
                                                    icon: <Trash2 className="size-4" />,
                                                    label: "Delete",
                                                    danger: true,
                                                    onClick: () => setToDelete(g),
                                                    disabledReason: g.artifacts > 0 ? `Frozen: ${fmt(g.artifacts)} artifacts reference this game.` : undefined,
                                                },
                                            ]}
                                        />
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </BrowsePanel>

            <Drawer
                open={drawer?.mode === "new" || drawer?.mode === "edit"}
                onClose={() => setDrawer(null)}
                title={drawer?.mode === "edit" ? "Edit game" : "Add game"}
                subtitle={drawer?.mode === "edit" ? drawer.game.name : "Register a curated game."}
            >
                {drawer?.mode === "new" ? <GameForm onDone={() => setDrawer(null)} /> : null}
                {drawer?.mode === "edit" ? <GameForm editing={drawer.game} onDone={() => setDrawer(null)} /> : null}
            </Drawer>

            <Drawer
                open={drawer?.mode === "scrape"}
                onClose={() => setDrawer(null)}
                title="Enqueue scrape"
                subtitle={drawer?.mode === "scrape" ? drawer.game.name : undefined}
            >
                {drawer?.mode === "scrape" ? <ScrapeForm presetGameId={drawer.game.id} /> : null}
            </Drawer>

            <Drawer
                open={drawer?.mode === "images"}
                onClose={() => setDrawer(null)}
                title="Upload images"
                subtitle={drawer?.mode === "images" ? drawer.game.name : undefined}
            >
                {drawer?.mode === "images" ? <ImageUploadForm presetGameId={drawer.game.id} /> : null}
            </Drawer>

            <ConfirmDialog
                open={toDelete !== null}
                title={`Delete ${toDelete?.name}?`}
                body="This removes the game registration. It has no artifacts, so nothing else is affected."
                confirmLabel="Delete"
                pending={del.isPending}
                error={del.error instanceof ApiError ? del.error.message : null}
                onConfirm={confirmDelete}
                onCancel={cancelDelete}
            />
        </>
    );
}
