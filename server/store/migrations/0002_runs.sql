-- Anonymous players and server-side runs (OpenSpec change: runs).
CREATE TABLE players (
  id         TEXT    PRIMARY KEY,
  token_hash BLOB    NOT NULL UNIQUE, -- sha256 of the secret token; the token itself is never stored
  created_at INTEGER NOT NULL
) STRICT;

-- One row per run. state: ready | picking | over.
CREATE TABLE runs (
  id                TEXT    PRIMARY KEY,
  player_id         TEXT    NOT NULL REFERENCES players (id),
  name              TEXT    NOT NULL,
  state             TEXT    NOT NULL,
  match_no          INTEGER NOT NULL, -- next / current match, 1-based
  wins              INTEGER NOT NULL,
  score             INTEGER NOT NULL,
  veterans          BLOB,             -- chill.v1.Army protobuf (units only), in army order
  current_battle_id TEXT,             -- started for match_no, not yet resolved
  end_reason        TEXT,
  created_at        INTEGER NOT NULL,
  ended_at          INTEGER
) STRICT;

CREATE INDEX runs_player ON runs (player_id, created_at);
CREATE INDEX runs_board ON runs (state, score DESC, ended_at);

ALTER TABLE battles ADD COLUMN run_id TEXT REFERENCES runs (id);
ALTER TABLE battles ADD COLUMN match_no INTEGER;
