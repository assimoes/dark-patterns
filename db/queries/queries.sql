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


-- name: AddIndividual :exec
INSERT INTO individuals (population_id, artifact_id)
VALUES ($1, $2)
ON CONFLICT (population_id, artifact_id) DO NOTHING;

-- name: UpsertAnnotation :one
-- Idempotent. The WHERE guard means an already completed annotation is not touched
-- Returning yields no rows in this case, and the worker treats it as already done
INSERT INTO annotations (run_id, individual_id, annotator_id, status, raw_response)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (run_id, individual_id, annotator_id) DO UPDATE
    SET status = EXCLUDED.status,
        raw_response = EXCLUDED.raw_response
    WHERE annotations.status <> 'completed'
RETURNING id;

-- name: ListUnannotatedTextReviews :many
-- The annotation worker's queue: items in the run's population not yet successfully annotated by an annotator
SELECT
    i.id as population_item_id,
    a.id as artifact_id,
    trd.body
FROM runs r
JOIN individuals i ON i.population_id = r.population_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN text_review_details trd ON trd.artifact_id = a.id
WHERE r.id = $1
    AND NOT EXISTS (
        SELECT 1 FROM annotations an
        WHERE an.run_id = r.id
            AND an.individual_id = i.id
            AND an.annotator_id = $2
            AND an.status = 'completed'
    )
ORDER BY i.id
LIMIT $3;


-- name: UpsertScrapeCursor :exec
INSERT INTO scrape_cursors (external_game_id, filter, language, cursor, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (external_game_id, filter, language) DO UPDATE
    SET cursor = EXCLUDED.cursor,
        updated_at = now();

-- name: GetScrapeCursor :one
SELECT cursor FROM scrape_cursors
WHERE external_game_id = $1 AND filter = $2 AND language = $3;