-- name: ListMesoPatternsByVersion :many
-- The taxonomy as of a pinned version
SELECT
m.id,
m.code,
m.name,
m.description,
h.name as high_level_pattern,
h.code as high_level_code,
m.examples,
m.counter_examples,
m.gray_mapping,
m.source_mapping
FROM taxonomy_meso_levels m
JOIN taxonomy_high_levels h ON h.id = m.parent_id
WHERE m.version = sqlc.arg(version)
ORDER BY h.id, m.code;

-- name: ListHighLevelsForMesoVersion :many
-- The strategic-intent parents used by a pinned meso version
SELECT DISTINCT h.id, h.code, h.name, coalesce(h.definition, h.description) AS definition
FROM taxonomy_high_levels h
JOIN taxonomy_meso_levels m ON m.parent_id = h.id
WHERE m.version = sqlc.arg(version)
ORDER BY h.id;

-- name: DeleteAnnotationPatterns :exec
-- Clear prior pattern rows before re-writing, so a retry doesn't leave stale ones.
DELETE FROM annotation_patterns WHERE annotation_id = $1;

-- name: InsertAnnotationPattern :exec
INSERT INTO annotation_patterns (annotation_id, pattern_id, evidence, explanation, confidence)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (annotation_id, pattern_id) DO NOTHING;