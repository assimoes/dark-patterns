ALTER TABLE game_display ALTER COLUMN external_game_id DROP IDENTITY IF EXISTS;
ALTER TABLE game_display DROP COLUMN source_refs;
