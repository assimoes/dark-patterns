-- name: GetTextReviewForIndividual :one
-- text-specific query
SELECT a.id AS artifact_id, td.body
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
JOIN text_review_details td on td.artifact_id = a.id
WHERE i.id = $1;


-- name: ListUnnanotatedIndividuals :many
-- The work queue for one panel member.
-- Modality agnostic
SELECT
    i.id as individual_id
FROM runs r
JOIN individuals i ON i.population_id = r.population_id
WHERE r.id = sqlc.arg(run_id)
    AND NOT EXISTS (
        SELECT 1 FROM annotations an
        WHERE an.run_id = r.id
            AND an.individual_id = i.id
            AND an.annotator_id = sqlc.arg(annotator_id)
            AND an.status = 'completed' 
    )
ORDER BY i.id;