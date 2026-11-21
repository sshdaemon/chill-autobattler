-- Classic (faction 1: warrior, archer, mage) is retired (OpenSpec change
-- retire-classic): the sim no longer knows it, so its unfinished runs end.
-- Their scores stay on the board, as do finished Classic runs. The DEFAULT 1
-- on runs.faction and battles.blue_faction (0006) stays: SQLite can't change
-- a column default in place, and the server always writes the faction.
UPDATE runs
SET state = 'over', end_reason = 'faction retired', current_battle_id = NULL,
    ended_at = CAST(strftime('%s', 'now') AS INTEGER) * 1000
WHERE state != 'over' AND faction = 1;
