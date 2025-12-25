-- 2025_01_internal_player_id.sql
-- Introduces internal player ID (not yet enforced as PK)

-- ============================================================
-- UP MIGRATION
-- Convert players to use internal integer PK
-- ============================================================

BEGIN;

-- 1. Add internal ID to players
ALTER TABLE players
ADD COLUMN id SERIAL;

-- 2. Ensure internal ID is unique (pre-PK)
ALTER TABLE players
ADD CONSTRAINT players_id_unique UNIQUE (id);

-- 3. Migrate users table to reference internal ID
ALTER TABLE users
ADD COLUMN player_internal_id INT;

UPDATE users u
SET player_internal_id = p.id
FROM players p
WHERE u.player_id = p.player_id;

ALTER TABLE users
ADD CONSTRAINT users_player_internal_id_fkey
FOREIGN KEY (player_internal_id) REFERENCES players(id);

ALTER TABLE users
DROP CONSTRAINT users_player_id_fkey;

ALTER TABLE users
DROP COLUMN player_id;

ALTER TABLE users
RENAME COLUMN player_internal_id TO player_id;

-- 4. Migrate gameplay_checkpoints (if present)
ALTER TABLE gameplay_checkpoints
ADD COLUMN player_internal_id INT;

UPDATE gameplay_checkpoints g
SET player_internal_id = p.id
FROM players p
WHERE g.player_id = p.player_id;

ALTER TABLE gameplay_checkpoints
ADD CONSTRAINT gameplay_checkpoints_player_internal_id_fkey
FOREIGN KEY (player_internal_id) REFERENCES players(id);

ALTER TABLE gameplay_checkpoints
DROP COLUMN player_id;

ALTER TABLE gameplay_checkpoints
RENAME COLUMN player_internal_id TO player_id;

-- 5. Drop old primary key on external ID
ALTER TABLE players
DROP CONSTRAINT players_pkey;

-- 6. Promote internal ID to primary key
ALTER TABLE players
ADD CONSTRAINT players_pkey PRIMARY KEY (id);

-- 7. Preserve external ID as unique identifier
ALTER TABLE players
ADD CONSTRAINT players_player_id_unique UNIQUE (player_id);

COMMIT;

-- ============================================================
-- DOWN MIGRATION
-- Revert back to external player_id as PK
-- ============================================================

BEGIN;

-- 1. Restore external-ID FKs
ALTER TABLE users
ADD COLUMN player_external_id TEXT;

UPDATE users u
SET player_external_id = p.player_id
FROM players p
WHERE u.player_id = p.id;

ALTER TABLE gameplay_checkpoints
ADD COLUMN player_external_id TEXT;

UPDATE gameplay_checkpoints g
SET player_external_id = p.player_id
FROM players p
WHERE g.player_id = p.id;

-- 2. Drop internal-ID FKs
ALTER TABLE users
DROP CONSTRAINT users_player_internal_id_fkey;

ALTER TABLE gameplay_checkpoints
DROP CONSTRAINT gameplay_checkpoints_player_internal_id_fkey;

-- 3. Remove internal PK
ALTER TABLE players
DROP CONSTRAINT players_pkey;

-- 4. Restore external PK
ALTER TABLE players
ADD CONSTRAINT players_pkey PRIMARY KEY (player_id);

-- 5. Remove internal references
ALTER TABLE users
DROP COLUMN player_id;

ALTER TABLE users
RENAME COLUMN player_external_id TO player_id;

ALTER TABLE gameplay_checkpoints
DROP COLUMN player_id;

ALTER TABLE gameplay_checkpoints
RENAME COLUMN player_external_id TO player_id;

-- 6. Remove internal ID column
ALTER TABLE players
DROP CONSTRAINT players_id_unique;

ALTER TABLE players
DROP COLUMN id;

-- 7. Restore FK constraints
ALTER TABLE users
ADD CONSTRAINT users_player_id_fkey
FOREIGN KEY (player_id) REFERENCES players(player_id);

ALTER TABLE gameplay_checkpoints
ADD CONSTRAINT gameplay_checkpoints_player_id_fkey
FOREIGN KEY (player_id) REFERENCES players(player_id);

COMMIT;
