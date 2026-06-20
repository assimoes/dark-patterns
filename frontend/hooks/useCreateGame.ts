import { api } from "@/lib/api";
import { qk } from "@/lib/queryKeys";
import { CreateGameInput } from "@/lib/types";
import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useCreateGame() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (body: CreateGameInput) => api.operations.createGame(body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: qk.dashboard })
            queryClient.invalidateQueries({ queryKey: qk.games })
        }
    })
}