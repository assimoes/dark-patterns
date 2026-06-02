CREATE INDEX ON individuals(population_id);
CREATE INDEX ON individuals(artifact_id);
CREATE INDEX ON annotations(run_id);
CREATE INDEX ON annotations(individual_id);
CREATE INDEX ON annotations(annotator_id);
CREATE INDEX ON annotation_patterns(annotation_id);
CREATE INDEX ON annotation_patterns(pattern_id);
CREATE INDEX ON taxonomy_meso_levels(parent_id);
