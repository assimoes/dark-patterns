-- a review can be adjudicated two ways against the same gold run: open (the auditor sees the panel)
-- and blind (the panel is hidden). pass keeps both as separate rows so one never overwrites the other.
-- existing rows are the open pass, which is what the default backfills.
ALTER TABLE adjudications
    ADD COLUMN pass varchar(8) NOT NULL DEFAULT 'open'
    CHECK (pass IN ('open', 'blind'));

-- the old key let one cell hold a single label per run. widen it by pass so the blind label sits beside
-- the open one instead of conflicting with it.
ALTER TABLE adjudications
    DROP CONSTRAINT adjudications_run_id_individual_id_pattern_id_key;

ALTER TABLE adjudications
    ADD CONSTRAINT adjudications_run_individual_pattern_pass_key
    UNIQUE (run_id, individual_id, pattern_id, pass);
