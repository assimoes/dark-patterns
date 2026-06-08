CREATE TABLE IF NOT EXISTS adjudications (
    id bigint generated always as identity primary key,
    run_id int not null references runs(id),
    individual_id bigint not null references individuals(id),
    pattern_id int not null references taxonomy_meso_levels(id),

    final_label boolean not null,
    direction varchar(12) not null check (direction in ('confirmation', 'replacement')),
    adjudicator_id int not null references annotators(id),
    panel_seed_at_adjudication jsonb not null,
    decided_at timestamptz not null default now(),
    unique (run_id, individual_id, pattern_id)
);

CREATE INDEX IF NOT EXISTS idx_adjudications_run_pattern ON adjudications(run_id, pattern_id);