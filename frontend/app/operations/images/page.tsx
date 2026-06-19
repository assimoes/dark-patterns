"use client";

import { useState } from "react";
import { Upload } from "lucide-react";
import { Card } from "@/components/ui/Card";
import {
    ErrorPanel,
    Field,
    SubmitButton,
    SuccessPanel,
    inputClass,
} from "@/components/operations/form";
import { GameSelect } from "@/components/operations/GameSelect";
import { useUploadImages } from "@/hooks/useUploadImages";
import { ApiError } from "@/lib/api/utils";
import type { UploadImagesInput } from "@/lib/types";

export default function ImagesPage() {
    const [gameId, setGameId] = useState("");
    const [files, setFiles] = useState<File[]>([]);
    // optional provenance note per file, keyed by its index in the picked list.
    const [descriptions, setDescriptions] = useState<Record<number, string>>({});

    const uploadImages = useUploadImages();

    const onPick = (list: FileList | null) => {
        setFiles(list ? Array.from(list) : []);
        setDescriptions({});
    };

    const setDesc = (i: number, v: string) =>
        setDescriptions((prev) => ({ ...prev, [i]: v }));

    const onSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!gameId || files.length === 0) return;
        const body: UploadImagesInput = {
            game_id: Number(gameId),
            images: files.map((file, i) => ({ file, description: descriptions[i] || "" })),
        };
        uploadImages.mutate(body);
    };

    const errorMessage =
        uploadImages.error instanceof ApiError
            ? uploadImages.error.message || "Request failed."
            : uploadImages.isError
                ? "Could not upload the images. Please try again."
                : null;

    return (
        <Card
            icon={<Upload className="size-5" />}
            iconClass="bg-violet-50 text-violet-600"
            title="Upload images"
            subtitle="Add screenshots to a game as image artifacts."
        >
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
                <GameSelect value={gameId} onChange={setGameId} required />

                <Field label="Image files" hint="PNG or JPEG screenshots. Pick one or more.">
                    <input
                        type="file"
                        multiple
                        accept="image/*"
                        onChange={(e) => onPick(e.target.files)}
                        className={inputClass}
                    />
                </Field>

                {files.length > 0 ? (
                    <div className="flex flex-col gap-3">
                        {files.map((f, i) => (
                            <Field
                                key={`${f.name}-${i}`}
                                label={f.name}
                                hint="Optional note: where it came from, why it shows a pattern. Kept as provenance, not shown to the model or in the blind pass."
                            >
                                <input
                                    type="text"
                                    value={descriptions[i] ?? ""}
                                    onChange={(e) => setDesc(i, e.target.value)}
                                    className={inputClass}
                                    placeholder="description (optional)"
                                />
                            </Field>
                        ))}
                    </div>
                ) : null}

                <ErrorPanel message={errorMessage} />

                <SubmitButton
                    pending={uploadImages.isPending}
                    idleLabel="Upload images"
                    pendingLabel="Uploading…"
                />

                {uploadImages.data ? (
                    <SuccessPanel title="Images uploaded" data={uploadImages.data} />
                ) : null}
            </form>
        </Card>
    );
}