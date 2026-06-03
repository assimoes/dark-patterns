ALTER TABLE annotations DROP COLUMN IF EXISTS response_meta;
DROP TABLE IF EXISTS run_annotators;
ALTER TABLE runs DROP COLUMN IF EXISTS config_digest;
ALTER TABLE runs DROP COLUMN IF EXISTS taxonomy_version;
