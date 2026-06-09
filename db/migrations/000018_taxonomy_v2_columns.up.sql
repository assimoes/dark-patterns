ALTER TABLE taxonomy_meso_levels
    ADD COLUMN examples text[],
    ADD COLUMN counter_examples text[],
    ADD COLUMN gray_mapping text[],
    ADD COLUMN source_mapping text[];

ALTER TABLE taxonomy_high_levels ADD COLUMN code varchar(2);

UPDATE taxonomy_high_levels SET code = 'TM' WHERE name = 'Temporal Manipulation';
UPDATE taxonomy_high_levels SET code = 'PM' WHERE name = 'Predatory Monetization';
UPDATE taxonomy_high_levels SET code = 'SE' WHERE name = 'Social Exploitation';
UPDATE taxonomy_high_levels SET code = 'PE' WHERE name = 'Psychological Exploitation';
UPDATE taxonomy_high_levels SET code = 'DR' WHERE name = 'Deceptive Representation';

ALTER TABLE taxonomy_high_levels ALTER COLUMN code SET NOT NULL;

ALTER TABLE annotation_patterns ADD COLUMN confidence numeric;
