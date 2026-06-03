ALTER TABLE runs ADD COLUMN taxonomy_version int;
ALTER TABLE runs ADD COLUMN config_digest text;

CREATE TABLE IF NOT EXISTS run_annotators (
    run_id int not null references runs(id) on delete cascade,
    annotator_id int not null references annotators(id),
    provider text not null,
    model_slug text not null,
    client_version text not null,
    sampling jsonb not null,
    primary key (run_id, annotator_id)
);

ALTER TABLE annotations ADD COLUMN response_meta jsonb;
