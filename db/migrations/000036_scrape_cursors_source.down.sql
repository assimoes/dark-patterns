ALTER TABLE scrape_cursors DROP CONSTRAINT scrape_cursors_pkey;
ALTER TABLE scrape_cursors ADD PRIMARY KEY (external_game_id, filter, language);
ALTER TABLE scrape_cursors DROP COLUMN source;
