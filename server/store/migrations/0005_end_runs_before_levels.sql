-- Sim 0.6.0 caps units at three abilities (levels 1 to 3), so veterans saved
-- by older servers may now be invalid. Unfinished runs end; their scores stay
-- on the board.
UPDATE runs
SET state = 'over', end_reason = 'rules changed', current_battle_id = NULL,
    ended_at = CAST(strftime('%s', 'now') AS INTEGER) * 1000
WHERE state != 'over';
