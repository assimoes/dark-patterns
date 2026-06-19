"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import type { UploadImagesInput } from "@/lib/types";

// Stores uploaded screenshots as image artifacts. invalidates the dashboard so the new images show in
// the game counts.
export function useUploadImages() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (body: UploadImagesInput) => api.operations.uploadImages(body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.dashboard });
        },
    });
}