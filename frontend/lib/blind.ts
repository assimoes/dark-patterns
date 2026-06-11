import { adjRun } from "@/lib/run";

// A blind review is the panel-free projection of a run review
export type BlindReview = {
    id: number;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
};

export const blindReviews: BlindReview[] = adjRun.reviews.map(
    ({ id, gameId, votedUp, language, body }) => ({ id, gameId, votedUp, language, body }),
);