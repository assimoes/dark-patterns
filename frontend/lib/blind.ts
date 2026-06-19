// maps the API blind read (BlindReviewData) onto the BlindReview shape the review pane renders
import type { BlindReviewData } from "@/lib/types";

// A blind review is the panel-free, answer-free projection of a run review: the review text and its
// display metadata, with both the panel votes and the auditor's saved gold labels stripped off. It is
// what the pane shows during a blind pass, where the labeller decides from the codebook alone.
export type BlindReview = {
    id: string;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
    modality?: string;
    imageUri?: string;
};

// projects the blind endpoint payload onto BlindReview. goldLabels are dropped here on purpose: only the
// worklist keeps them (to pre-fill the cards on revisit); they must never reach the display surface.
// mirrors mapAdjudicationReview for the open pass.
export function mapBlindReview(data: BlindReviewData): BlindReview {
    return {
        id: data.id,
        gameId: data.gameId,
        votedUp: data.votedUp,
        language: data.language,
        body: data.body,
        modality: data.modality,
        imageUri: data.imageUri,
    };
}
