-- name: SnapshotRunAnnotator :exec
-- Freeze one panel member
INSERT INTO run_annotators (run_id, annotator_id, provider, model_slug, client_version, sampling)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (run_id, annotator_id) DO NOTHING;

-- name: ListRunAnnotators :many
-- Read the frozen panel.
SELECT annotator_id, provider, model_slug, client_version, sampling
FROM run_annotators
WHERE run_id = $1
ORDER BY annotator_id;

-- name: SetRunConfigDigest :exec
-- Stamp the digest once. A second snapshot is a no-op
UPDATE runs SET config_digest = $2 WHERE id = $1 AND config_digest IS NULL;