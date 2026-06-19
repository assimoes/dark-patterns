ALTER TABLE text_review_details
    ADD COLUMN voted_up boolean not null default false,
    ADD COLUMN hours_played int not null default 0,
    ADD COLUMN weighted_vote_score numeric;

UPDATE text_review_details
SET voted_up = coalesce((source_meta->>'voted_up')::boolean, false),
    hours_played = coalesce((source_meta->>'hours_played')::int, 0),
    weighted_vote_score = (source_meta->>'weighted_vote_score')::numeric;
