-- 2025_02_fix_players_primary_key.sql
-- Enforces internal player ID as primary key and fixes FKs

-- up
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_pkey;
ALTER TABLE players ADD CONSTRAINT players_pkey PRIMARY KEY (id);

-- down
ALTER TABLE players DROP CONSTRAINT players_pkey;
ALTER TABLE players ADD CONSTRAINT players_pkey PRIMARY KEY (player_id);
