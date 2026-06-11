"use client";

import { useState } from "react";
import { Gamepad2 } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    Field,
    SubmitButton,
    SuccessPanel,
    inputClass,
} from "@/components/operations/form";
import { useCreateGame } from "@/hooks/useCreateGame";
import { ApiError } from "@/lib/api";
import type { CreateGameInput, Monetization } from "@/lib/types";

export default function GamesPage() {
    const [externalGameId, setExternalGameId] = useState("");
    const [name, setName] = useState("");
    const [short, setShort] = useState("");
    const [monetization, setMonetization] = useState<"" | Monetization>("");
    const [color, setColor] = useState("#6366f1");

    const createGame = useCreateGame();

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const body: CreateGameInput = {
            external_game_id: Number(externalGameId),
        };
        if (name.trim()) body.name = name.trim();
        if (short.trim()) body.short = short.trim();
        if (monetization) body.monetization = monetization;
        if (color) body.color = color;
        createGame.mutate(body);
    };

    const errorMessage =
        createGame.error instanceof ApiError
            ? createGame.error.message || "Request failed."
            : createGame.isError
                ? "Could not add the game. Please try again."
                : null;

    return (
        <Card
            icon={<Gamepad2 className="size-5" />}
            iconClass="bg-amber-50 text-amber-600"
            title="Add game"
            subtitle="Register a curated Steam game."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
                <Field
                    label="External game id"
                    hint="The numeric Steam app id. Required."
                >
                    <input
                        type="number"
                        required
                        value={externalGameId}
                        onChange={(e) => setExternalGameId(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Name">
                    <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Short label">
                    <input
                        type="text"
                        value={short}
                        onChange={(e) => setShort(e.target.value)}
                        className={inputClass}
                    />
                </Field>

                <Field label="Monetization">
                    <select
                        value={monetization}
                        onChange={(e) =>
                            setMonetization(e.target.value as "" | Monetization)
                        }
                        className={inputClass}
                    >
                        <option value="">— default —</option>
                        <option value="f2p">f2p</option>
                        <option value="premium">premium</option>
                    </select>
                </Field>

                <Field label="Accent color">
                    <input
                        type="color"
                        value={color}
                        onChange={(e) => setColor(e.target.value)}
                        className="h-10 w-16 cursor-pointer rounded-lg border border-slate-200 bg-white p-1"
                    />
                </Field>

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={createGame.isPending}
                    idleLabel="Add game"
                    pendingLabel="Adding…"
                />

                {createGame.data ? (
                    <SuccessPanel title="Game created" data={createGame.data} />
                ) : null}
            </form>
        </Card>
    );
}