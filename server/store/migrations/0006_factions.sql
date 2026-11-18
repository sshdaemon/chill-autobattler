-- The faction a run plays (sim.Faction; 1 = classic), and per battle the
-- faction Blue's army was checked against, so old battles keep verifying with
-- the value they were issued with (OpenSpec change: faction-pick).
ALTER TABLE runs ADD COLUMN faction INTEGER NOT NULL DEFAULT 1;
ALTER TABLE battles ADD COLUMN blue_faction INTEGER NOT NULL DEFAULT 1;
