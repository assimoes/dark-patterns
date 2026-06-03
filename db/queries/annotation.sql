-- name: ListMesoPatternsByVersion :many
-- The taxonomy as of a pinned version
SELECT
m.id,
m.code,
m.name,
m.description,
h.name as high_level_pattern
FROM taxonomy_meso_levels m
JOIN taxonomy_high_levels h ON h.id = m.parent_id
WHERE m.version = sqlc.arg(version)
ORDER BY m.code;

-- name: DeleteAnnotationPatterns :exec
-- Clear prior pattern rows before re-writing, so a retry doesn't leave stale ones.
DELETE FROM annotation_patterns WHERE annotation_id = $1;

-- name: InsertAnnotationPattern :exec
INSERT INTO annotation_patterns (annotation_id, pattern_id, evidence, explanation)
VALUES ($1, $2, $3, $4)
ON CONFLICT (annotation_id, pattern_id) DO NOTHING;