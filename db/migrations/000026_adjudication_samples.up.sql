CREATE TABLE IF NOT EXISTS adjudication_samples (
    id bigint generated always as identity primary key,
    panel_run_id int not null references runs(id),
    gold_run_id int not null references runs(id),
    strategy text not null,
    seed bigint not null,
    params jsonb not null,
    created_at timestamptz not null default now()
);

CREATE TABLE IF NOT EXISTS adjudication_sample_items (
    sample_id bigint not null references adjudication_samples(id) on delete cascade,
    individual_id bigint not null references individuals(id),
    external_game_id int not null,
    stratum text not null check (stratum in ('flagged_majority','flagged_split','silent')),
    selection_prob double precision not null,
    primary key (sample_id, individual_id)
);
 
CREATE INDEX IF NOT EXISTS idx_adj_sample_items_sample ON adjudication_sample_items(sample_id);