-- Drop old tables
DROP TABLE IF EXISTS gameplay_checkpoints;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS users;

-- One per Descope login
CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    descope_id  TEXT UNIQUE NOT NULL,
    username    TEXT UNIQUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Up to 10 per user
CREATE TABLE players (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    playername  TEXT UNIQUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Many per player
CREATE TABLE gameplay_checkpoints (
    id          SERIAL PRIMARY KEY,
    player_id   INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    data        JSONB NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);