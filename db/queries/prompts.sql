-- name: CreatePrompt :one
INSERT INTO prompts (name, version, modality, system_prompt, template)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetPrompt :one
SELECT * FROM prompts WHERE id = $1;

-- name: GetPromptByNameVersion :one
SELECT * FROM prompts where name = $1 AND version = $2;

-- name: GetLatestPrompt :one
SELECT * FROM prompts WHERE name = $1 ORDER BY version DESC LIMIT 1;