-- going back to the single-label-per-cell key. blind rows would collide with the open ones under it,
-- so they get dropped first. the open pass is the original data and stays.
DELETE FROM adjudications WHERE pass = 'blind';

ALTER TABLE adjudications
    DROP CONSTRAINT adjudications_run_individual_pattern_pass_key;

ALTER TABLE adjudications
    ADD CONSTRAINT adjudications_run_id_individual_id_pattern_id_key
    UNIQUE (run_id, individual_id, pattern_id);

ALTER TABLE adjudications DROP COLUMN pass;
