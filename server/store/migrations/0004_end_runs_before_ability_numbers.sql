-- Sim 0.5.0 numbers abilities per unit (1..5) instead of with global skill
-- ids, so veterans saved by older servers carry ids that now mean other
-- abilities (or none). Unfinished runs end; their scores stay on the board.
UPDATE runs
SET state = 'over', end_reason = 'rules changed', current_battle_id = NULL,
    ended_at = CAST(strftime('%s', 'now') AS INTEGER) * 1000
WHERE state != 'over';
