-- name: UpdateGameDisplay :one
-- edit a games display fields and its per-source handles.
UPDATE game_display
SET name          = $2,
    short         = $3,
    monetization  = $4,
    display_color = $5,
    source_refs   = $6
WHERE external_game_id = $1
RETURNING external_game_id, name, short, monetization, display_color, source_refs;

-- name: GetGameSourceRef :one
-- the scrape handle a game registered for one source (e.g. its subreddit for 'reddit'). empty when the
-- game has no binding for that source.
SELECT COALESCE(source_refs->>sqlc.arg(source)::text, '')::text AS handle
FROM game_display
WHERE external_game_id = sqlc.arg(external_game_id);

-- name: CountArtifactsForGame :one
SELECT count(*)::int FROM artifacts WHERE external_game_id = $1;

-- name: GameArtifactTotals :many
-- artifacts per game, the delete guard: a game with any artifact is frozen.
SELECT external_game_id, count(*)::int AS artifacts
FROM artifacts
GROUP BY external_game_id;

-- name: DeleteGameDisplay :exec
DELETE FROM game_display WHERE external_game_id = $1;

-- name: UpdatePrompt :exec
UPDATE prompts
SET name = $2, version = $3, modality = $4, system_prompt = $5, template = $6
WHERE id = $1;

-- name: CountRunsForPrompt :one
SELECT count(*)::int FROM runs WHERE prompt_id = $1;

-- name: DeletePrompt :exec
DELETE FROM prompts WHERE id = $1;

-- name: UpdateAnnotatorLabel :exec
UPDATE annotators SET label = $2 WHERE id = $1;

-- name: CountAnnotatorRefs :one
SELECT
    (SELECT count(*) FROM run_annotators ra WHERE ra.annotator_id = $1)::int +
    (SELECT count(*) FROM annotations an WHERE an.annotator_id = $1)::int AS ref_count;

-- name: DeleteAnnotator :exec
DELETE FROM annotators WHERE id = $1;

-- name: RunImpact :one
-- the blast radius of deleting a run: its annotations, the adjudication samples it seeds (as panel or
-- gold), and its adjudications.
SELECT
    (SELECT count(*) FROM annotations an WHERE an.run_id = $1)::int AS annotations,
    (SELECT count(*) FROM adjudication_samples s WHERE s.panel_run_id = $1 OR s.gold_run_id = $1)::int AS samples,
    (SELECT count(*) FROM adjudications adj WHERE adj.run_id = $1)::int AS adjudications;

-- name: PopulationImpact :one
-- the blast radius of deleting a population: its individuals, the runs over it, and those runs
-- annotations and adjudication samples.
SELECT
    (SELECT count(*) FROM individuals i WHERE i.population_id = $1)::int AS individuals,
    (SELECT count(*) FROM runs r WHERE r.population_id = $1)::int AS runs,
    (SELECT count(*) FROM annotations an JOIN runs r ON r.id = an.run_id WHERE r.population_id = $1)::int AS annotations,
    (SELECT count(*) FROM adjudication_samples s JOIN runs r ON r.id = s.panel_run_id WHERE r.population_id = $1)::int AS samples;

-- name: ListRunIDsByPopulation :many
SELECT id FROM runs WHERE population_id = $1;

-- name: DeleteAnnotationPatternsByRun :exec
DELETE FROM annotation_patterns
WHERE annotation_id IN (SELECT id FROM annotations WHERE run_id = $1);

-- name: DeleteAnnotationsByRun :exec
DELETE FROM annotations WHERE run_id = $1;

-- name: DeleteAdjudicationSamplesByRun :exec
-- sample items cascade on the sample delete.
DELETE FROM adjudication_samples WHERE panel_run_id = $1 OR gold_run_id = $1;

-- name: DeleteAdjudicationsByRun :exec
DELETE FROM adjudications WHERE run_id = $1;

-- name: DeleteRun :exec
-- run_annotators cascade on the run delete.
DELETE FROM runs WHERE id = $1;

-- name: DeleteIndividualsByPopulation :exec
DELETE FROM individuals WHERE population_id = $1;

-- name: DeletePopulation :exec
DELETE FROM populations WHERE id = $1;
