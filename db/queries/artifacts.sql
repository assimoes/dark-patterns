-- name: UpsertArtifact :one
-- Idempotent on source_id; DO UPDATE instead of DO NOTHING because we want it to always return the id even on re-scrape.
-- ON CONFLICT we update the scraped_at date.
INSERT INTO artifacts (modality, source, source_id, content_hash, scraped_at, external_game_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (source, source_id) DO UPDATE
    SET scraped_at = EXCLUDED.scraped_at
RETURNING id;

-- name: UpsertTextReviewDetail :exec
-- On re-scrape we only refresh the volatile signal (votes, hours).
-- body, lang and score stay as first seen on purpose, so annotation always line up with the text they ran on.
-- An edited review is a new artifact if we ever want to recapture it.
INSERT INTO text_review_details (artifact_id, body, voted_up, hours_played, lang, weighted_vote_score)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (artifact_id) DO UPDATE
    SET voted_up = EXCLUDED.voted_up,
        hours_played = EXCLUDED.hours_played;

-- name: GetArtifact :one
SELECT * FROM artifacts WHERE id = $1;

-- name: CountArtifactsBySource :one
SELECT COUNT(*) FROM artifacts WHERE source = $1 AND modality = $2;

-- name: UpsertImageDetail :exec
INSERT INTO image_details (artifact_id, image_uri, width, height, mime_type, ocr_text, description)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING;
