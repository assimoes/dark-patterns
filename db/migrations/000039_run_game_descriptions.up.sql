-- the per-run freeze of which approved description each game used, mirroring run_annotators for the
-- panel. a run annotates a population that spans many games, so the run pins one description per game.
-- their content hashes fold into config_digest, and the annotation loader reads the pinned version.
CREATE TABLE IF NOT EXISTS run_game_descriptions (
    run_id int not null references runs(id),
    external_game_id int not null,
    game_description_id int not null references game_descriptions(id),
    primary key (run_id, external_game_id)
);
