-- name: NextDescriptionVersion :one
-- the next version number for a games description chain (1 when it has none).
SELECT COALESCE(max(version), 0)::int + 1 AS version
FROM game_descriptions
WHERE external_game_id = $1;

-- name: InsertGameDescription :one
-- store a researched draft (or an invalid/error draft the reviewer must fix).
INSERT INTO game_descriptions (
    external_game_id, version, status, profile, rendered_text, research_model, sources, valence_flags, error
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetGameDescription :one
SELECT * FROM game_descriptions WHERE id = $1;

-- name: ListGameDescriptionsForGame :many
-- newest first; the review console lists the version history of one game.
SELECT * FROM game_descriptions
WHERE external_game_id = $1
ORDER BY version DESC;

-- name: ListGameDescriptionStates :many
-- the latest version's status per game, for the games list badge.
SELECT DISTINCT ON (external_game_id) external_game_id, status
FROM game_descriptions
ORDER BY external_game_id, version DESC;

-- name: GetLatestApprovedDescription :one
SELECT * FROM game_descriptions
WHERE external_game_id = $1 AND status = 'approved';

-- name: UpdateDraftDescription :one
-- edit a draft in place: the reviewers fixes to the structured profile and the rendered text, plus the
-- re-run valence flags. only drafts are editable.
UPDATE game_descriptions
SET profile = $2, rendered_text = $3, sources = $4, valence_flags = $5
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: ApproveDescription :one
-- freeze a draft as the approved version. blocked unless it is a draft with no unresolved valence flags.
UPDATE game_descriptions
SET status = 'approved', approved_by = $2, approved_at = now()
WHERE id = $1 AND status = 'draft' AND valence_flags = '[]'::jsonb
RETURNING *;

-- name: SupersedePriorApproved :exec
-- demote the current approved description for a game (called before approving a newer version).
UPDATE game_descriptions
SET status = 'superseded'
WHERE external_game_id = $1 AND status = 'approved';

-- name: GamesMissingApprovedDescription :many
-- the run gate: distinct games present in a population that have no approved description yet.
SELECT DISTINCT a.external_game_id, gd.name
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
JOIN game_display gd ON gd.external_game_id = a.external_game_id
WHERE i.population_id = $1
    AND NOT EXISTS (
        SELECT 1 FROM game_descriptions d
        WHERE d.external_game_id = a.external_game_id AND d.status = 'approved'
    )
ORDER BY a.external_game_id;

-- name: PinRunGameDescriptions :exec
-- freeze the approved description per game in the runs population onto the run.
INSERT INTO run_game_descriptions (run_id, external_game_id, game_description_id)
SELECT sqlc.arg(run_id), a.external_game_id, d.id
FROM (
    SELECT DISTINCT a.external_game_id
    FROM individuals i
    JOIN artifacts a ON a.id = i.artifact_id
    WHERE i.population_id = sqlc.arg(population_id)
) a
JOIN game_descriptions d ON d.external_game_id = a.external_game_id AND d.status = 'approved'
ON CONFLICT (run_id, external_game_id) DO NOTHING;

-- name: ListRunGameDescriptions :many
-- the pinned descriptions for a run: game id, version, content hash, and the text to inject.
SELECT rgd.external_game_id, d.version, d.content_hash, d.rendered_text
FROM run_game_descriptions rgd
JOIN game_descriptions d ON d.id = rgd.game_description_id
WHERE rgd.run_id = $1
ORDER BY rgd.external_game_id;
