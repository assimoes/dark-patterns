CREATE TABLE IF NOT EXISTS populations (
    id int generated always as identity primary key,
    modality varchar(20) not null check (modality in ('text', 'image', 'video')),
    description text,
    criteria jsonb,
    artifacts_cutoff timestamptz not null,
    created_at timestamptz not null default now()
);

CREATE TABLE IF NOT EXISTS individuals (
    id bigint generated always as identity primary key,
    population_id int not null references populations(id),
    artifact_id bigint not null references artifacts(id),
    created_at timestamptz not null default now(),
    unique(population_id, artifact_id)
);