CREATE TABLE IF NOT EXISTS models (
    id int generated always as identity primary key,
    family varchar(50) not null,
    slug varchar(150) not null,
    name varchar(150) not null,
    modalities varchar(20)[] not null default '{text}',
    active boolean not null default true,
    created_at timestamptz not null default now(),
    unique(slug)
);

CREATE TABLE IF NOT EXISTS annotators (
    id int generated always as identity primary key,
    kind varchar(10) not null check (kind in ('llm', 'human')),
    model_id int references models(id),
    label varchar(100) not null,
    created_at timestamptz not null default now(),
    check ((kind = 'llm') = (model_id is not null))
);

CREATE TABLE IF NOT EXISTS prompts (
    id int generated always as identity primary key,
    name varchar(50) not null,
    version int not null,
    modality varchar(20) not null check (modality in ('text', 'image', 'video')),
    system_prompt text,
    template text not null,
    content_hash bytea generated always as (
        digest (coalesce(system_prompt, '') || '|' || template, 'sha256')
    ) stored,
    created_at timestamptz not null default now(),
    unique(name, version)
);

CREATE TABLE IF NOT EXISTS runs (
    id int generated always as identity primary key,
    run_type varchar(10) not null check (run_type in ('llm_panel', 'gold')),
    population_id int not null references populations(id),
    prompt_id int not null references prompts(id),
    temperature numeric,
    top_p numeric,
    params jsonb,
    created_at timestamptz not null default now()
);

