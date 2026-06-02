CREATE TABLE IF NOT EXISTS scrape_cursors (
    external_game_id int not null,
    filter varchar(10) not null,
    language varchar(25) not null,
    cursor text not null,
    updated_at timestamptz not null default now(),
    primary key (external_game_id, filter, language)
);