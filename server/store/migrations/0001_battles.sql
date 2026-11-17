-- One row per issued battle. Lifecycle: started -> locked -> verified.
-- Seeds are uint64 stored as the same 64 bits in a signed INTEGER.
-- Times are Unix milliseconds.
CREATE TABLE IF NOT EXISTS battles (
  id              TEXT    PRIMARY KEY,
  sim_version     TEXT    NOT NULL,
  map_seed        INTEGER NOT NULL,
  combat_seed     INTEGER NOT NULL,
  red_skill_count INTEGER NOT NULL,
  created_at      INTEGER NOT NULL,
  deploy_deadline INTEGER NOT NULL,
  blue_army       BLOB,    -- chill.v1.Army protobuf, set by LockDeployment
  locked_at       INTEGER,
  verify_deadline INTEGER,
  verified_at     INTEGER,
  fair            INTEGER, -- 0/1
  reason          TEXT,
  outcome         BLOB     -- chill.v1.BattleOutcome protobuf (the server's)
) STRICT;
