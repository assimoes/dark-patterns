-- name: UpsertArtifact :one
-- idempotent on source_id; DO UPDATE because we want it to always return the id even on re-scrape.
-- ON CONFLICT we update the scraped_at date.
INSERT INTO artifacts (modality, source, source_id, content_hash, scraped_at, external_game_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (source, source_id) DO UPDATE
    SET scraped_at = EXCLUDED.scraped_at
RETURNING id;

-- name: UpsertTextReviewDetail :exec
-- text channel: body and lang are universal, source_meta holds everything source-specific
-- (Steam voted_up/hours/score, Reddit subreddit/post_id). on re-scrape we refresh source_meta.
INSERT INTO text_review_details (artifact_id, body, lang, source_meta)
VALUES ($1, $2, $3, $4)
ON CONFLICT (artifact_id) DO UPDATE
    SET source_meta = EXCLUDED.source_meta;

-- name: GetArtifact :one
SELECT * FROM artifacts WHERE id = $1;

-- name: CountArtifactsBySource :one
SELECT COUNT(*) FROM artifacts WHERE source = $1 AND modality = $2;

-- name: UpsertImageDetail :exec
INSERT INTO image_details (artifact_id, image_uri, width, height, mime_type, ocr_text, description)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING;
