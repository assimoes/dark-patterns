-- a saved comparison of two runs over the same population: the experiment's declared contract (which run
-- attributes must be equal, which must differ) plus the two run ids. the review-by-review divergence is
-- computed from annotations on demand, not stored.
CREATE TABLE IF NOT EXISTS comparisons (
    id int generated always as identity primary key,
    label text not null,
    run_a_id int not null references runs(id),
    run_b_id int not null references runs(id),
    contract jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);
