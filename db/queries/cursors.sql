-- name: UpsertScrapeCursor :exec
INSERT INTO scrape_cursors (external_game_id, filter, language, cursor, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (external_game_id, filter, language) DO UPDATE
    SET cursor = EXCLUDED.cursor,
        updated_at = now();

-- name: GetScrapeCursor :one
SELECT cursor FROM scrape_cursors
WHERE external_game_id = $1 AND filter = $2 AND language = $3;
