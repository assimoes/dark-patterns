-- name: UpsertScrapeCursor :exec
INSERT INTO scrape_cursors (external_game_id, source, filter, language, cursor, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (external_game_id, source, filter, language) DO UPDATE
    SET cursor = EXCLUDED.cursor,
        updated_at = now();

-- name: GetScrapeCursor :one
SELECT cursor FROM scrape_cursors
WHERE external_game_id = $1 AND source = $2 AND filter = $3 AND language = $4;
