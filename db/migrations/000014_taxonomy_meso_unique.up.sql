 
ALTER TABLE taxonomy_meso_levels
    DROP CONSTRAINT IF EXISTS taxonomy_meso_levels_code_key;

ALTER TABLE taxonomy_meso_levels
    ADD CONSTRAINT taxonomy_meso_levels_code_version_key UNIQUE (code, version);
