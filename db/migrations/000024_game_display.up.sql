-- The frontend renders a name, a short code, a monetization badge and an accent colour per game.
-- The pipeline never needed those, so artifacts only carries external_game_id. Rather than hardcode
-- a lookup in Go, we keep that presentation metadata in the database: one row per game keyed by the
-- same external_game_id artifacts already use. It is a side table (not new columns on artifacts)
-- because artifacts holds many review rows per game and this is one fact per game.
CREATE TABLE IF NOT EXISTS game_display (
    external_game_id int primary key,
    name text not null,
    short text not null,
    monetization text not null check (monetization in ('f2p', 'b2p', 'sub')),
    display_color text not null
);

INSERT INTO game_display (external_game_id, name, short, monetization, display_color)
  VALUES
      (230410,  'Warframe',          'Warframe', 'f2p', '#6366f1'),
      (552990,  'World of Warships', 'WWS',      'f2p', '#33ab12'),
      (1407200, 'World of Tanks',    'WWT',      'f2p', '#7d70db'),
      (2694490, 'Path of Exile II',  'PoE 2',    'f2p', '#c12f84'),
      (3405690, 'FIFA 2026',         'FC26',     'f2p', '#d01616')
  ON CONFLICT (external_game_id) DO NOTHING;