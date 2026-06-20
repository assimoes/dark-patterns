-- name: ListGameDisplays :many
-- the curated games with their presentation metadata and per-source handles. external_game_id is the
-- stable internal id the frontend uses as `gameId`; source_refs maps each source to its scrape handle.
SELECT external_game_id, name, short, monetization, display_color, source_refs
FROM game_display
ORDER BY external_game_id;
 
-- name: CountIndividualsPerGame :many
-- curated individuals (reviews in a population) per game, across every population.
SELECT a.external_game_id, count(*)::int AS individuals
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
GROUP BY a.external_game_id
ORDER BY a.external_game_id;
 
-- name: ListRunsForDashboard :many
-- one row per run with its populations modality as a human label and its creation time.
-- annotator_ids is the panel the run pinned; members are resolved separately per run.
SELECT
    r.id,
    r.run_type,
    p.modality AS population,
    r.created_at,
    r.annotator_ids
FROM runs r
JOIN populations p ON p.id = r.population_id
ORDER BY r.created_at DESC;
 
-- name: ListMembersForRun :many
-- the panel members of a run: every annotator referenced by runs.annotator_ids, with the kind
-- (llm | human) and a display label. LLM members carry their model name, humans their own label.
SELECT
    an.id,
    an.kind,
    COALESCE(m.name, an.label) AS label
FROM runs r
JOIN annotators an ON an.id = ANY(r.annotator_ids)
LEFT JOIN models m ON m.id = an.model_id
WHERE r.id = sqlc.arg(run_id)
ORDER BY an.kind, an.id;
 
-- name: ReviewStatsPerGame :many
SELECT
    a.external_game_id,
    r.id AS run_id,
    p.name AS prompt,
    count(DISTINCT i.id)::int AS reviews,
    count(DISTINCT i.id) FILTER (WHERE an.status = 'completed')::int AS annotated
FROM runs r
JOIN individuals i ON i.population_id = r.population_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN prompts p ON p.id = r.prompt_id
LEFT JOIN annotations an ON an.individual_id = i.id AND an.run_id = r.id
GROUP BY a.external_game_id, r.id, p.name
ORDER BY a.external_game_id, r.id;


 
-- name: GetPatternIDByCode :one
-- resolve a meso pattern code (e.g. 'PM-1') the frontend sends to its row id, within a taxonomy
-- version. code is unique only per (code, version), so the version is required or the wrong versions
-- id comes back — which would make a saved adjudication unreadable against the runs actual taxonomy.
SELECT id FROM taxonomy_meso_levels WHERE code = sqlc.arg(code) AND version = sqlc.arg(version);
 
-- name: GetPanelRunForReview :one
-- the llm panel run whose votes seed a decision on this review. one panel per population; pick the
-- most recent so the frozen seed reflects the panel the auditor is actually looking at.
SELECT r.id
FROM runs r
JOIN individuals i ON i.population_id = r.population_id
WHERE i.id = sqlc.arg(individual_id)
    AND r.run_type = 'llm_panel'
ORDER BY r.created_at DESC
LIMIT 1;

-- name: ListPopulationsForGame :many
-- the populations that contain this games reviews, each with the games slice: how many of the
-- games individuals fall in the population, and how many of those have a completed annotation. a
-- population is multi-game (stratified, per-game capped), so this is THIS games part of it. counts
-- are per population, never summed across them.
SELECT
    p.id AS population_id,
    p.description,
    p.created_at,
    count(DISTINCT i.id)::int AS reviews,
    count(DISTINCT i.id) FILTER (
        WHERE EXISTS (
            SELECT 1 FROM annotations an
            WHERE an.individual_id = i.id AND an.status = 'completed'
        )
    )::int AS annotated
FROM populations p
JOIN individuals i ON i.population_id = p.id
JOIN artifacts a ON a.id = i.artifact_id
WHERE a.external_game_id = sqlc.arg(external_game_id)
GROUP BY p.id, p.description, p.created_at
ORDER BY p.id;
 
-- name: ModelStatsForGamePopulation :many
-- per LLM model, the number of DISTINCT reviews of one game annotated within one population. counting
-- distinct individuals (not annotation rows) and scoping to a single population makes the models
-- comparable: inside one population they all share the same work set, so a complete run shows every
-- model at the populations slice size, not a runaway sum across runs.
SELECT
    m.name AS model,
    count(DISTINCT i.id)::int AS annotated
FROM annotations an
JOIN individuals i ON i.id = an.individual_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN annotators ann ON ann.id = an.annotator_id
JOIN models m ON m.id = ann.model_id
WHERE a.external_game_id = sqlc.arg(external_game_id)
    AND i.population_id = sqlc.arg(population_id)
    AND an.status = 'completed'
    AND ann.kind = 'llm'
GROUP BY m.name
ORDER BY annotated DESC, m.name;
 
-- name: PanelForPopulation :many
-- the panel that worked a population: every annotator frozen onto any of the populations runs, with
-- its kind (llm | human) and a display label (the model name for an llm, the annotators own label
-- for a human). run_annotators is the single source of "who annotates this run".
SELECT DISTINCT
    an.kind,
    COALESCE(m.name, an.label) AS label
FROM run_annotators ra
JOIN runs r ON r.id = ra.run_id
JOIN annotators an ON an.id = ra.annotator_id
LEFT JOIN models m ON m.id = an.model_id
WHERE r.population_id = sqlc.arg(population_id)
ORDER BY an.kind, label;

-- name: InsertGameDisplay :one
-- register a game so it appears on the dashboard list and can be scraped. external_game_id is left to
-- the sequence (a high internal id), source_refs holds the per-source scrape handles.
INSERT INTO game_display (name, short, monetization, display_color, source_refs)
VALUES ($1, $2, $3, $4, $5)
RETURNING external_game_id, name, short, monetization, display_color, source_refs;

-- name: ListPopulations :many
-- every population with its size (the number of frozen individuals). LEFT JOIN so an empty population
-- still appears with a zero count. newest first, the order an operator picking a population wants.
SELECT
    p.id,
    p.modality,
    p.description,
    p.created_at,
    count(i.id)::int AS individuals
FROM populations p
LEFT JOIN individuals i ON i.population_id = p.id
GROUP BY p.id
ORDER BY p.id DESC;
 
-- name: GetPopulation :one
-- one population row by id, for the population/run detail headers (modality, description, created_at).
SELECT id, modality, description, created_at FROM populations WHERE id = sqlc.arg(id);
 
-- name: PopulationPerGame :many
-- for one population, its per-game slice: how many of the populations reviews belong to each game and
-- how many of those have a completed annotation. mirrors ListPopulationsForGame but pivots to group by
-- game within a single population instead of by population within a single game.
SELECT
    a.external_game_id,
    count(DISTINCT i.id)::int AS reviews,
    count(DISTINCT i.id) FILTER (
        WHERE EXISTS (
            SELECT 1 FROM annotations an
            WHERE an.individual_id = i.id AND an.status = 'completed'
        )
    )::int AS annotated
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
WHERE i.population_id = sqlc.arg(population_id)
GROUP BY a.external_game_id
ORDER BY a.external_game_id;
 
-- name: ListPrompts :many
-- every prompt as a form option: its id, name, version and modality, plus how many runs use it (a
-- prompt with runs is frozen). ordered by id for a stable list.
SELECT p.id, p.name, p.version, p.modality,
    (SELECT count(*) FROM runs r WHERE r.prompt_id = p.id)::int AS run_count
FROM prompts p
ORDER BY p.id;
 
-- name: ListRuns :many
-- every run with what an operator needs to recognise and pick it: its type, the population modality as a
-- label, the foreign keys, the taxonomy version, when it ran, and the size of the panel it pinned.
SELECT
    r.id,
    r.run_type,
    p.modality AS population,
    r.population_id,
    r.prompt_id,
    r.taxonomy_version,
    r.created_at,
    coalesce(array_length(r.annotator_ids, 1), 0)::int AS panel_size
FROM runs r
JOIN populations p ON p.id = r.population_id
ORDER BY r.created_at DESC;
 
-- name: ListAnnotatorsWithModel :many
-- every annotator as a form option, carrying the display model name for llm annotators (NULL for
-- humans). ListAnnotators returns the raw rows with only a model_id; this resolves the name in SQL so
-- the API never has to look models up one by one.
SELECT
    an.id,
    an.kind,
    an.label,
    m.name AS model,
    (SELECT count(*) FROM run_annotators ra WHERE ra.annotator_id = an.id)::int +
    (SELECT count(*) FROM annotations a WHERE a.annotator_id = an.id)::int AS ref_count
FROM annotators an
LEFT JOIN models m ON m.id = an.model_id
ORDER BY an.kind, an.label;

-- name: GameReviewTotals :many
SELECT
    a.external_game_id,
    count(DISTINCT i.id)::int AS reviews,
    count(DISTINCT i.id) FILTER (
        WHERE EXISTS (
            SELECT 1 FROM annotations an
            WHERE an.individual_id = i.id AND an.status = 'completed'
        )
    )::int AS annotated
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
GROUP BY a.external_game_id
ORDER BY a.external_game_id;

-- name: ModelStatsForGameRun :many
SELECT
    m.name AS model,
    count(DISTINCT i.id)::int AS annotated
FROM annotations an
JOIN individuals i ON i.id = an.individual_id
JOIN artifacts a ON a.id = i.artifact_id
JOIN annotators ann ON ann.id = an.annotator_id
JOIN models m ON m.id = ann.model_id
WHERE a.external_game_id = sqlc.arg(external_game_id)
    AND an.run_id = sqlc.arg(run_id)
    AND an.status = 'completed'
    AND ann.kind = 'llm'
GROUP BY m.name
ORDER BY annotated DESC, m.name;

-- name: GetGoldRunForPopulation :one
-- the most recent gold run for a population. adjudication samples and decisions write into one gold
-- run per population; pick the newest so a freshly drawn sample lands on the run the auditor reads.
SELECT id
FROM runs
WHERE population_id = sqlc.arg(population_id)
    AND run_type = 'gold'
ORDER BY created_at DESC
LIMIT 1;

-- name: ClassifyReviewsForSampling :many
-- the candidate pool for a sample: every completed review in the population, tagged with its stratum.
-- a review is classified by the panels per-pattern votes: flagged_majority if any pattern reached a
-- majority Present (n_present*2 > n_total), else flagged_split if any pattern had Present votes without
-- a majority (the panel disagreed), else silent (no pattern got a single Present vote).
WITH completed AS (
    SELECT an.individual_id, count(*)::int AS n_total
    FROM annotations an WHERE an.run_id = sqlc.arg(panel_run_id) AND an.status = 'completed'
    GROUP BY an.individual_id
),
cell AS (
    SELECT an.individual_id, ap.pattern_id, count(*)::int AS n_present
    FROM annotations an JOIN annotation_patterns ap ON ap.annotation_id = an.id
    WHERE an.run_id = sqlc.arg(panel_run_id) AND an.status = 'completed'
    GROUP BY an.individual_id, ap.pattern_id
),
verdict AS (
    SELECT c.individual_id,
        bool_or(c.n_present * 2 > t.n_total) AS any_majority,
        bool_or(c.n_present > 0 AND c.n_present * 2 <= t.n_total) AS any_split
    FROM cell c JOIN completed t ON t.individual_id = c.individual_id
    GROUP BY c.individual_id
)
SELECT i.id AS individual_id, a.external_game_id,
    (CASE WHEN COALESCE(v.any_majority, false) THEN 'flagged_majority'
          WHEN COALESCE(v.any_split, false) THEN 'flagged_split'
          ELSE 'silent' END)::text AS stratum
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
JOIN completed t ON t.individual_id = i.id
LEFT JOIN verdict v ON v.individual_id = i.id
WHERE i.population_id = sqlc.arg(population_id)
ORDER BY a.external_game_id, i.id;
 

-- name: InsertAdjudicationSample :one
-- the sample header. params holds the per-stratum target Ns; seed is stored so the draw is auditable.
INSERT INTO adjudication_samples (panel_run_id, gold_run_id, strategy, seed, params)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;
 
-- name: InsertAdjudicationSampleItem :exec
-- one frozen member of a sample: its stratum and its inverse-probability weight (drawn / stratum_size).
INSERT INTO adjudication_sample_items (sample_id, individual_id, external_game_id, stratum, selection_prob)
VALUES ($1, $2, $3, $4, $5);
 
-- name: GetLatestSampleForRun :one
-- the most recent sample drawn for a panel run, to reopen its queue.
SELECT id, panel_run_id, gold_run_id, strategy, seed, params, created_at
FROM adjudication_samples
WHERE panel_run_id = sqlc.arg(panel_run_id)
ORDER BY created_at DESC
LIMIT 1;
 
-- name: ListSampleReviews :many
-- the queue for a sample
SELECT
    it.individual_id,
    it.external_game_id,
    it.stratum,
    a.modality,
    COALESCE((td.source_meta->>'voted_up')::boolean, false)::boolean AS voted_up,
    COALESCE(td.lang, '')::text AS lang,
    (
        SELECT count(*)::int
        FROM adjudications adj
        WHERE adj.run_id = s.gold_run_id
            AND adj.individual_id = it.individual_id
            AND adj.pass = sqlc.arg(pass)
    ) AS decided
FROM adjudication_sample_items it
JOIN adjudication_samples s ON s.id = it.sample_id
JOIN individuals i ON i.id = it.individual_id
JOIN artifacts a ON a.id = i.artifact_id
LEFT JOIN text_review_details td ON td.artifact_id = a.id
WHERE it.sample_id = sqlc.arg(sample_id)
ORDER BY it.external_game_id, it.individual_id;

-- name: GetReviewMeta :one
-- the review header for the auditor/blind views
SELECT a.modality,
       COALESCE(td.body, '')::text AS body,
       COALESCE((td.source_meta->>'voted_up')::boolean, false)::boolean AS voted_up,
       COALESCE(td.lang, '')::text AS lang,
       COALESCE(img.image_uri, '')::text AS image_uri,
       COALESCE(img.description, '')::text AS description,
       a.external_game_id
FROM individuals i
JOIN artifacts a ON a.id = i.artifact_id
LEFT JOIN text_review_details td ON td.artifact_id = a.id
LEFT JOIN image_details img ON img.artifact_id = a.id
WHERE i.id = sqlc.arg(individual_id);