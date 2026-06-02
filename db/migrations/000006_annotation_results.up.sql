CREATE TABLE IF NOT EXISTS annotations (
    id bigint generated always as identity primary key,
    run_id int not null references runs(id),
    individual_id bigint not null references individuals(id),
    annotator_id int not null references annotators(id),
    status varchar(12) not null default 'completed' check (status in ('completed', 'failed', 'parse_error')),
    raw_response jsonb,
    created_at timestamptz not null default now(),
    unique (run_id, individual_id, annotator_id)
);

CREATE TABLE IF NOT EXISTS annotation_patterns (
    id bigint generated always as identity primary key,
    annotation_id bigint not null references annotations(id),
    pattern_id int not null references taxonomy_meso_levels(id),
    evidence text,
    explanation text,
    created_at timestamptz not null default now(),
    unique(annotation_id, pattern_id)
);