ALTER TABLE annotation_patterns DROP COLUMN IF EXISTS confidence;

ALTER TABLE taxonomy_high_levels DROP COLUMN IF EXISTS code;

ALTER TABLE taxonomy_meso_levels
    DROP COLUMN IF EXISTS examples,
    DROP COLUMN IF EXISTS counter_examples,
    DROP COLUMN IF EXISTS gray_mapping,
    DROP COLUMN IF EXISTS source_mapping;
