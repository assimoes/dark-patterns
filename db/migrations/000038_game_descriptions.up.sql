-- a per-game neutral description, produced by a research call and frozen on human approval. it is a
-- versioned immutable artifact like prompts: editing an approved one creates a new version. annotation
-- injects the approved version's rendered_text so the model knows what words in a review refer to.
CREATE TABLE IF NOT EXISTS game_descriptions (
    id int generated always as identity primary key,
    external_game_id int not null references game_display(external_game_id),
    version int not null,
    status text not null check (status in ('draft', 'approved', 'superseded', 'invalid', 'error')),
    profile jsonb not null default '{}'::jsonb,
    rendered_text text not null default '',
    research_model text not null default '',
    sources jsonb not null default '[]'::jsonb,
    valence_flags jsonb not null default '[]'::jsonb,
    error text not null default '',
    content_hash bytea generated always as (digest(rendered_text, 'sha256')) stored,
    created_at timestamptz not null default now(),
    approved_by int,
    approved_at timestamptz,
    unique (external_game_id, version)
);

-- at most one approved description per game; the rest are draft/superseded history.
CREATE UNIQUE INDEX game_descriptions_one_approved
    ON game_descriptions (external_game_id)
    WHERE status = 'approved';
