-- name: CreateRun :one
INSERT INTO runs (run_type, population_id, prompt_id, temperature, top_p, params)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetRun :one
SELECT * FROM runs WHERE id = $1;

-- name: ListRunsByPopulation :many
SELECT * FROM runs WHERE population_id = $1 ORDER BY created_at DESC;

