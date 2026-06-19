-- steam columns move into source_meta so text_review_details is just (artifact_id, body,
-- lang, source_meta) for every text source.
UPDATE text_review_details
SET source_meta = coalesce(source_meta, '{}'::jsonb) || jsonb_build_object(
        'voted_up', voted_up,
        'hours_played', hours_played,
        'weighted_vote_score', weighted_vote_score
    )
WHERE source_meta IS NULL OR source_meta = '{}'::jsonb;

ALTER TABLE text_review_details
    DROP COLUMN voted_up,
    DROP COLUMN hours_played,
    DROP COLUMN weighted_vote_score;
