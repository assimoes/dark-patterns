-- text_review_details is the text channel for every text source now. source-specific metadata goes
-- in source_meta
ALTER TABLE text_review_details ADD COLUMN source_meta jsonb;
