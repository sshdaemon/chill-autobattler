-- name: CreateBattle :exec
INSERT INTO battles (id, sim_version, map_seed, combat_seed, red_skill_count, created_at, deploy_deadline)
VALUES (@id, @sim_version, @map_seed, @combat_seed, @red_skill_count, @created_at, @deploy_deadline);

-- name: GetBattle :one
SELECT * FROM battles WHERE id = @id;

-- Succeeds (1 row) only for an unlocked battle before its deploy deadline.
-- name: LockBattle :execrows
UPDATE battles
SET blue_army = @blue_army, locked_at = @now, verify_deadline = @verify_deadline
WHERE id = @id AND locked_at IS NULL AND deploy_deadline > @now;

-- Succeeds (1 row) only for a locked, not yet verified battle before its verify deadline.
-- name: RecordVerdict :execrows
UPDATE battles
SET verified_at = @now, fair = @fair, reason = @reason, outcome = @outcome
WHERE id = @id AND locked_at IS NOT NULL AND verified_at IS NULL AND verify_deadline > @now;

-- ---------- players ----------

-- name: CreatePlayer :exec
INSERT INTO players (id, token_hash, created_at) VALUES (@id, @token_hash, @created_at);

-- name: PlayerByTokenHash :one
SELECT id FROM players WHERE token_hash = @token_hash;

-- ---------- runs ----------

-- name: CreateRun :exec
INSERT INTO runs (id, player_id, name, state, match_no, wins, score, created_at, faction)
VALUES (@id, @player_id, @name, 'ready', 1, 0, 0, @created_at, @faction);

-- name: GetRun :one
SELECT * FROM runs WHERE id = @id;

-- The player's active run if any, else their latest.
-- name: LatestRun :one
SELECT * FROM runs WHERE player_id = @player_id
ORDER BY (state != 'over') DESC, created_at DESC, id DESC LIMIT 1;

-- name: EndRun :exec
UPDATE runs SET state = 'over', end_reason = @end_reason, ended_at = @now, current_battle_id = NULL
WHERE id = @id AND state != 'over';

-- name: EndActiveRuns :exec
UPDATE runs SET state = 'over', end_reason = @end_reason, ended_at = @now, current_battle_id = NULL
WHERE player_id = @player_id AND state != 'over';

-- name: SetCurrentBattle :execrows
UPDATE runs SET current_battle_id = @battle_id
WHERE id = @id AND state = 'ready' AND current_battle_id IS NULL;

-- name: WinMatch :exec
UPDATE runs SET state = @state, wins = wins + 1, score = score + @points, match_no = match_no + 1,
  veterans = @veterans, current_battle_id = NULL
WHERE id = @id;

-- name: FinishPicks :execrows
UPDATE runs SET state = 'ready', veterans = @veterans WHERE id = @id AND state = 'picking';

-- Ends every active run whose current battle missed its deadline.
-- name: SweepAbandoned :exec
UPDATE runs SET state = 'over', end_reason = 'abandoned', ended_at = @now, current_battle_id = NULL
WHERE state != 'over' AND current_battle_id IN (
  SELECT b.id FROM battles b
  WHERE (b.locked_at IS NULL AND b.deploy_deadline <= @now)
     OR (b.locked_at IS NOT NULL AND b.verified_at IS NULL AND b.verify_deadline <= @now)
);

-- name: TopRuns :many
SELECT name, score, wins, player_id = @player_id AS mine FROM runs
WHERE state = 'over' ORDER BY score DESC, ended_at ASC, id ASC LIMIT 10;

-- name: BestFinishedScore :one
SELECT CAST(COALESCE(MAX(score), -1) AS INTEGER) FROM runs WHERE player_id = @player_id AND state = 'over';

-- name: CountFinishedAbove :one
SELECT COUNT(*) FROM runs WHERE state = 'over' AND score > @score;

-- ---------- battles in runs ----------

-- name: CreateRunBattle :exec
INSERT INTO battles (id, sim_version, map_seed, combat_seed, red_skill_count, created_at, deploy_deadline, run_id, match_no, red_health_pct, blue_faction)
VALUES (@id, @sim_version, @map_seed, @combat_seed, @red_skill_count, @created_at, @deploy_deadline, @run_id, @match_no, @red_health_pct, @blue_faction);
