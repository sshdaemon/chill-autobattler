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
