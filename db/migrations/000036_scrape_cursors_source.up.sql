-- scrape_cursors keys a saved page position per game. it gains a source so steam and reddit cursors for
-- the same game stay apart. existing rows backfill to steam.
ALTER TABLE scrape_cursors ADD COLUMN source text NOT NULL DEFAULT 'steam';

ALTER TABLE scrape_cursors DROP CONSTRAINT scrape_cursors_pkey;
ALTER TABLE scrape_cursors ADD PRIMARY KEY (external_game_id, source, filter, language);
