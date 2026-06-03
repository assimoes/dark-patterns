-- name: GetTextReviewForIndividual :one
-- text-specific query
SELECT a.id AS artifact_id, td.body
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
JOIN text_review_details td on td.artifact_id = a.id
WHERE i.id = $1;