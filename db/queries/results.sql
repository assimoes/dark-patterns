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
